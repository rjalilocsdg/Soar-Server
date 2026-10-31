package panel

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDowngradeReleaseSelectionAndReauthentication(t *testing.T) {
	b, _, h := testBrowser(t)
	fake := &fakeMaintenance{}
	h.SetMaintenance(fake, nil)
	releases := []PanelRelease{releaseFixture("0.14.1", false), releaseFixture("0.10.1", false), releaseFixture("0.9.0", false), releaseFixture(Version, false), releaseFixture("0.14.2-beta1", true)}
	draft := releaseFixture("0.14.0", false)
	draft.Draft = true
	releases = append(releases, draft)
	h.listPanelReleases = func(context.Context) ([]PanelRelease, error) { return releases, nil }
	b.request("/api/updates/downgrades", "POST", map[string]any{}, 401)
	b.login()
	raw := b.request("/api/updates/downgrades", "POST", map[string]any{}, 200)
	var result struct{ Versions []PanelDowngradeRelease }
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Versions) != 2 || result.Versions[0].Version != "0.14.1" {
		t.Fatalf("unexpected releases: %+v", result)
	}
	request := map[string]any{"version": "0.14.1", "password": "wrong", "confirm": true}
	b.request("/api/updates/downgrade", "POST", request, 403)
	request["password"] = "test-only-password-2026"
	request["confirm"] = false
	b.request("/api/updates/downgrade", "POST", request, 400)
	request["confirm"] = true
	for _, version := range []string{Version, "999.0.0", "0.9.0", "0.14.0-beta1", "https://evil.test"} {
		request["version"] = version
		b.request("/api/updates/downgrade", "POST", request, 400)
	}
	if len(fake.requests) != 0 {
		t.Fatal("invalid downgrade started")
	}
	request["version"] = "0.14.1"
	b.request("/api/updates/downgrade", "POST", request, 202)
	if len(fake.requests) != 1 || fake.requests[0].Action != "downgrade" {
		t.Fatal(fake.requests)
	}
	cfg := defaultEgressConfig()
	cfg.StormDNS.Mode = "socks5"
	if DowngradeCompatibility(cfg, "0.10.1") == nil || DowngradeCompatibility(cfg, "0.14.1") != nil {
		t.Fatal("proxy downgrade guard failed")
	}
}
