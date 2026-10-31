package panel

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

const MinimumPanelDowngrade = "0.10.1"

func ValidatePanelDowngrade(version, current string) error {
	if !ValidVersion(version) || strings.Contains(version, "-") || !NewerVersion(current, version) || NewerVersion(MinimumPanelDowngrade, version) {
		return errors.New("choose an older stable panel version, 0.10.1 or later")
	}
	return nil
}

func DowngradeCompatibility(cfg EgressConfig, version string) error {
	if NewerVersion("0.14.1", version) {
		for _, route := range []EngineEgress{cfg.Soar, cfg.Cotten, cfg.MasterDNS, cfg.StormDNS} {
			if route.proxy() {
				return errors.New("older panel cannot preserve configured proxy routes; choose 0.14.1 or later")
			}
		}
	}
	return nil
}

// Inspect a consistent private snapshot, never the live Bolt database. The
// privileged maintenance service repeats the route check just before activation.
func ValidateDowngradeSnapshot(path, version string) error {
	db, err := bolt.Open(path, 0600, &bolt.Options{ReadOnly: true, Timeout: time.Second})
	if err != nil {
		return err
	}
	defer db.Close()
	return db.View(func(tx *bolt.Tx) error {
		if tx.Bucket(meta) == nil {
			return errors.New("invalid panel database")
		}
		cfg, err := readEgress(tx)
		if err != nil {
			return err
		}
		return DowngradeCompatibility(cfg, version)
	})
}

type PanelDowngradeRelease struct {
	Version    string `json:"version"`
	URL        string `json:"release_url"`
	Compatible bool   `json:"compatible"`
	Reason     string `json:"reason,omitempty"`
}

func selectPanelDowngrades(releases []PanelRelease, current string, cfg EgressConfig) []PanelDowngradeRelease {
	out := []PanelDowngradeRelease{}
	seen := map[string]bool{}
	for _, r := range releases {
		v := strings.TrimPrefix(r.Tag, "panel-v")
		if r.Prerelease || seen[v] || ValidatePanelDowngrade(v, current) != nil || validatePanelRelease(r, v) != nil {
			continue
		}
		item := PanelDowngradeRelease{Version: v, URL: r.URL, Compatible: true}
		if err := DowngradeCompatibility(cfg, v); err != nil {
			item.Compatible = false
			item.Reason = err.Error()
		}
		out = append(out, item)
		seen[v] = true
	}
	sort.Slice(out, func(i, j int) bool { return NewerVersion(out[i].Version, out[j].Version) })
	return out
}

func (h *HTTP) panelDowngrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path == "/api/updates/downgrades" {
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()
		list := h.listPanelReleases
		if list == nil {
			list = ListPanelReleases
		}
		releases, err := list(ctx)
		if err != nil {
			errorReply(w, 502, err)
			return
		}
		cfg, err := h.store.Egress()
		if err != nil {
			errorReply(w, 500, err)
			return
		}
		jsonReply(w, 200, map[string]any{"current": Version, "versions": selectPanelDowngrades(releases, Version, cfg)})
		return
	}
	var in struct {
		Version  string `json:"version"`
		Password string `json:"password"`
		Confirm  bool   `json:"confirm"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		errorReply(w, 400, err)
		return
	}
	if !h.maintenancePassword(w, r, in.Password) {
		return
	}
	if !in.Confirm {
		errorReply(w, 400, errors.New("explicit downgrade confirmation required"))
		return
	}
	if err := ValidatePanelDowngrade(in.Version, Version); err != nil {
		errorReply(w, 400, err)
		return
	}
	cfg, err := h.store.Egress()
	if err != nil {
		errorReply(w, 500, err)
		return
	}
	if err = DowngradeCompatibility(cfg, in.Version); err != nil {
		errorReply(w, 409, err)
		return
	}
	if h.ops == nil {
		errorReply(w, 503, errors.New("maintenance service unavailable"))
		return
	}
	job, err := h.ops.Start(r.Context(), MaintenanceRequest{Action: "downgrade", Version: in.Version})
	if err != nil {
		errorReply(w, 409, err)
		return
	}
	_ = h.store.Audit("maintenance.downgrade_requested", "")
	jsonReply(w, 202, job)
}
