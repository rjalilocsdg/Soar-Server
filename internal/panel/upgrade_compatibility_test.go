package panel

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"os"
	"reflect"
	"testing"
)

// Golden accounts and links were exported by the unmodified beta3 fork. Keys
// belong to disposable test accounts, never to a deployed server.
func TestBeta3UpgradePreservesAccountsProfilesAndSubscriptions(t *testing.T) {
	raw, err := os.ReadFile("testdata/beta3-accounts.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Settings Settings `json:"settings"`
		Accounts []struct {
			Client json.RawMessage   `json:"client"`
			Seed   []byte            `json:"seed"`
			Token  string            `json:"token"`
			Links  map[string]string `json:"links"`
		} `json:"accounts"`
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	s := testStore(t)
	if err = s.db.Update(func(tx *bolt.Tx) error {
		if err := putJSON(tx.Bucket(meta), []byte("settings"), fixture.Settings); err != nil {
			return err
		}
		for _, entry := range fixture.Accounts {
			var c Client
			if err := json.Unmarshal(entry.Client, &c); err != nil {
				return err
			}
			if err := tx.Bucket(clients).Put([]byte(c.ID), entry.Client); err != nil {
				return err
			}
			if err := tx.Bucket(seeds).Put([]byte(c.ID), entry.Seed); err != nil {
				return err
			}
			token, err := base64.RawURLEncoding.DecodeString(entry.Token)
			if err != nil {
				return err
			}
			if err := tx.Bucket(subscriptions).Put([]byte(c.ID), token); err != nil {
				return err
			}
			if err := tx.Bucket(subscriptionIndex).Put(subscriptionDigest(token), []byte(c.ID)); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range fixture.Accounts {
		var original Client
		json.Unmarshal(entry.Client, &original)
		got, err := s.Get(original.ID)
		if err != nil {
			t.Fatal(err)
		}
		// New fields resolve defaults in memory without migrating stored accounts.
		original.MasterDNSSessionLimit = DefaultMasterDNSSessionLimit
		original.SoarVersion = DefaultSoarVersion
		if got != original {
			t.Fatalf("upgrade changed beta3 account: got=%+v want=%+v", got, original)
		}
		if !bytes.Equal(clientRecord(t, s, got.ID), entry.Client) {
			t.Fatal("reading rewrote account")
		}
		links, err := s.Profiles(got.ID)
		if err != nil || !reflect.DeepEqual(links, entry.Links) {
			t.Fatal("upgrade changed exported beta3 profiles", got.Engines, err)
		}
		seed, err := s.Seed(got.ID)
		if err != nil || !bytes.Equal(seed, entry.Seed) {
			t.Fatal("upgrade changed seed", err)
		}
		c, _, err := s.Subscription(entry.Token)
		if err != nil || c.ID != got.ID {
			t.Fatal("old subscription credential invalidated", err)
		}
		// Older API updates omit both the new session limit and encryption choice.
		in := input(got)
		in.Note = "metadata after upgrade"
		updated, err := s.Edit(got.ID, in)
		if err != nil || updated.EncryptionMethod != got.EncryptionMethod || updated.MasterDNSSessionLimit != 255 || updated.Zone != got.Zone || updated.PublicKey != got.PublicKey {
			t.Fatal("metadata edit lost account policy", updated, err)
		}
	}
	if err = s.db.View(func(tx *bolt.Tx) error { return inspectBackupTransaction(tx, &BackupSummary{}) }); err != nil {
		t.Fatal("upgraded backup failed validation", err)
	}
}
