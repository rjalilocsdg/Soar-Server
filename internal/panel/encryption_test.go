package panel

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"strings"
	"testing"
)

func TestEncryptionDefaultsPreservationProfilesAndBackup(t *testing.T) {
	s := testStore(t)
	c, err := s.Create(ClientInput{Name: "encryption", Engines: "all", Enabled: true})
	if err != nil || c.EncryptionMethod != 5 {
		t.Fatal(c, err)
	}
	// Old database records resolve to AES without being rewritten.
	var legacy map[string]any
	json.Unmarshal(clientRecord(t, s, c.ID), &legacy)
	delete(legacy, "encryption_method")
	if err = s.db.Update(func(tx *bolt.Tx) error { return putJSON(tx.Bucket(clients), []byte(c.ID), legacy) }); err != nil {
		t.Fatal(err)
	}
	before := clientRecord(t, s, c.ID)
	c, err = s.Get(c.ID)
	if err != nil || c.EncryptionMethod != 5 || !bytes.Equal(before, clientRecord(t, s, c.ID)) {
		t.Fatal("legacy record changed", err)
	}
	seed, _ := s.Seed(c.ID)
	originalLinks, _ := s.Profiles(c.ID)
	for _, method := range []int{1, 5} {
		in := input(c)
		in.EncryptionMethod = &method
		c, err = s.Edit(c.ID, in)
		if err != nil {
			t.Fatal(err)
		}
		in = input(c)
		in.Note = "older API edit"
		c, err = s.Edit(c.ID, in)
		if err != nil || c.EncryptionMethod != method {
			t.Fatal("omission reset encryption", c, err)
		}
		links, err := s.Profiles(c.ID)
		if err != nil {
			t.Fatal(err)
		}
		if links["soar"] != originalLinks["soar"] {
			t.Fatal("Soar profile changed")
		}
		for _, engine := range []string{"cotten", "stormdns", "masterdns"} {
			_, encoded, _ := strings.Cut(links[engine], "://")
			raw, err := base64.RawURLEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatal(err)
			}
			var body struct {
				Profile struct {
					Server struct {
						Method int `json:"encryption_method"`
					} `json:"server"`
				} `json:"profile"`
			}
			if err = json.Unmarshal(raw, &body); err != nil || body.Profile.Server.Method != method {
				t.Fatal("profile method mismatch", engine, string(raw), err)
			}
		}
		if err = s.db.View(func(tx *bolt.Tx) error { return inspectBackupTransaction(tx, &BackupSummary{}) }); err != nil {
			t.Fatal("backup rejected method", method, err)
		}
	}
	afterSeed, _ := s.Seed(c.ID)
	if !bytes.Equal(seed, afterSeed) {
		t.Fatal("encryption selection rotated key")
	}
	for _, method := range []int{-1, 0, 2, 3, 4, 6} {
		in := input(c)
		in.EncryptionMethod = &method
		if _, err = s.Edit(c.ID, in); err == nil {
			t.Fatal("accepted invalid method", method)
		}
	}
}

func TestEncryptionReconcileOnlyAffectedAccount(t *testing.T) {
	s := testStore(t)
	a := account(t, s, 0)
	b := account(t, s, 0)
	m := testManager(t, s, nil)
	before, other := m.accounts[a.ID], m.accounts[b.ID]
	in := input(a)
	in.EncryptionMethod = sessionLimitPointer(1)
	if _, err := s.Edit(a.ID, in); err != nil {
		t.Fatal(err)
	}
	if err := m.Reconcile(); err != nil {
		t.Fatal(err)
	}
	if m.accounts[a.ID] == before || before.ctx.Err() == nil || m.accounts[b.ID] != other || other.ctx.Err() != nil {
		t.Fatal("method change did not selectively retire runtime")
	}
}

func TestEncryptionHTTPValidation(t *testing.T) {
	b, s, _ := testBrowser(t)
	b.login()
	for _, value := range []any{0, 2, 3, 4, 6, -1, 1.5, "1"} {
		b.request("/api/clients", "POST", map[string]any{"name": "invalid", "enabled": true, "encryption_method": value}, 400)
	}
	c, err := s.Create(ClientInput{Name: "xor", Enabled: true, EncryptionMethod: sessionLimitPointer(1)})
	if err != nil {
		t.Fatal(err)
	}
	b.request("/api/clients/"+c.ID, "PUT", input(c), 200)
	c, err = s.Get(c.ID)
	if err != nil || c.EncryptionMethod != 1 {
		t.Fatal("HTTP omission reset method", c, err)
	}
}
