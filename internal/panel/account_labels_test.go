package panel

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func TestShortLabelsConcurrentAndFullAccountIdentity(t *testing.T) {
	s := testStore(t)
	var wg sync.WaitGroup
	created := make(chan Client, 96)
	for i := 0; i < cap(created); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := s.Create(ClientInput{Name: "short account", Enabled: true, Engines: "all"})
			if err != nil {
				t.Error(err)
				return
			}
			created <- c
		}()
	}
	wg.Wait()
	close(created)
	seen := map[string]bool{}
	for c := range created {
		label, _, _ := strings.Cut(c.Zone, ".")
		seed, err := s.Seed(c.ID)
		if len(label) != 3 || len(c.ID) != 32 || err != nil || len(seed) != 32 || seen[c.Zone] {
			t.Fatalf("invalid or duplicate identity: %+v, %v", c, err)
		}
		seen[c.Zone] = true
	}
	if len(seen) != cap(created) {
		t.Fatal("concurrent account allocation lost accounts")
	}
}

func TestAccountLabelCollisionRetryAndFailure(t *testing.T) {
	s := testStore(t)
	c := account(t, s, 0)
	cfg, _ := s.Settings()
	label, _, _ := strings.Cut(c.Zone, ".")
	replacement := "a" + label[1:]
	if replacement == label {
		replacement = "b" + label[1:]
	}
	attempts := 0
	err := s.db.Update(func(tx *bolt.Tx) error {
		zone, err := allocateAccountZone(tx, cfg, func(length int) (string, error) {
			attempts++
			if attempts == 1 {
				return label, nil
			}
			return replacement, nil
		})
		if err != nil || zone == c.Zone || attempts != 2 {
			t.Fatalf("collision retry: %q, %d, %v", zone, attempts, err)
		}
		_, err = allocateAccountZone(tx, cfg, func(int) (string, error) { return label, nil })
		if err == nil {
			t.Fatal("exhausted retries reused an occupied domain")
		}
		failure := errors.New("entropy unavailable")
		_, err = allocateAccountZone(tx, cfg, func(int) (string, error) { return "", failure })
		if !errors.Is(err, failure) {
			t.Fatal("random-source error was ignored")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if after, _ := s.Get(c.ID); after != c {
		t.Fatal("allocation attempt changed an existing account")
	}
}

func TestLongLabelsReserveTruncatedAliasPrefix(t *testing.T) {
	s := testStore(t)
	c := account(t, s, 0)
	cfg, _ := s.Settings()
	cfg.AccountLabelLength = 26
	if err := s.db.Update(func(tx *bolt.Tx) error {
		c.Zone = strings.Repeat("a", 26) + "." + cfg.BaseZone
		if err := putJSON(tx.Bucket(clients), []byte(c.ID), c); err != nil {
			return err
		}
		attempts := 0
		zone, err := allocateAccountZone(tx, cfg, func(length int) (string, error) {
			attempts++
			if attempts == 1 {
				return strings.Repeat("a", 20) + strings.Repeat("b", length-20), nil
			}
			return strings.Repeat("c", length), nil
		})
		if err != nil || attempts != 2 || strings.HasPrefix(zone, strings.Repeat("a", 20)) {
			t.Fatalf("alias collision was accepted: %q, %v", zone, err)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestLabelLengthPreservesExistingAccountsAndSupportsExplicitRotation(t *testing.T) {
	s := testStore(t)
	cfg, _ := s.Settings()
	cfg.AccountLabelLength = 26
	if err := s.SaveSettings(cfg); err != nil {
		t.Fatal(err)
	}
	c := account(t, s, 10000)
	seed, _ := s.Seed(c.ID)
	links, _ := s.Profiles(c.ID)
	token, _ := s.SubscriptionToken(c.ID, false)
	label, _, _ := strings.Cut(c.Zone, ".")
	if len(label) != 26 {
		t.Fatalf("configured label length ignored: %q", label)
	}
	cfg.AccountLabelLength = 3
	if err := s.SaveSettings(cfg); err != nil {
		t.Fatal(err)
	}
	after, _ := s.Get(c.ID)
	afterSeed, _ := s.Seed(c.ID)
	afterLinks, _ := s.Profiles(c.ID)
	afterToken, _ := s.SubscriptionToken(c.ID, false)
	if after != c || !bytes.Equal(seed, afterSeed) || !reflect.DeepEqual(links, afterLinks) || token != afterToken {
		t.Fatal("length setting changed an existing account or subscription")
	}
	rotated, err := s.Rotate(c.ID)
	if err != nil {
		t.Fatal(err)
	}
	newLabel, _, _ := strings.Cut(rotated.Zone, ".")
	if len(newLabel) != 3 || rotated.ID != c.ID || rotated.PublicKey == c.PublicKey || rotated.QuotaBytes != c.QuotaBytes {
		t.Fatalf("explicit rotation lost account identity or ignored short length: %+v", rotated)
	}
	oldAPI := cfg
	oldAPI.AccountLabelLength = 0
	if err := s.SaveSettings(oldAPI); err != nil {
		t.Fatal(err)
	}
	if latest, _ := s.Settings(); latest.AccountLabelLength != 3 {
		t.Fatal("older API reset label setting")
	}
}

func TestLabelLengthHTTPValidationAndOmission(t *testing.T) {
	b, s, _ := testBrowser(t)
	b.login()
	cfg, _ := s.Settings()
	for _, value := range []any{0, 2, 27, 3.5, "3"} {
		raw, _ := json.Marshal(cfg)
		var body map[string]any
		json.Unmarshal(raw, &body)
		body["account_label_length"] = value
		b.request("/api/settings", "PUT", body, 400)
	}
	cfg.AccountLabelLength = 12
	b.request("/api/settings", "PUT", cfg, 200)
	raw, _ := json.Marshal(cfg)
	var legacy map[string]any
	json.Unmarshal(raw, &legacy)
	delete(legacy, "account_label_length")
	b.request("/api/settings", "PUT", legacy, 200)
	if saved, _ := s.Settings(); saved.AccountLabelLength != 12 {
		t.Fatal("omitted field reset configured label length")
	}
	if _, err := s.Create(ClientInput{Name: "12 chars", Enabled: true}); err != nil {
		t.Fatal(err)
	}
}
