package panel

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	legacy "soar-v023"
	"testing"
	"time"
)

func TestSoarVersionSelectionPreservesIdentityAndTransfers(t *testing.T) {
	s := testStore(t)
	c := account(t, s, 0)
	seed, _ := s.Seed(c.ID)
	profiles, _ := s.Profiles(c.ID)
	m := testManager(t, s, func(ctx context.Context, network, address string) (net.Conn, error) {
		a, b := net.Pipe()
		go func() { defer b.Close(); io.Copy(b, b) }()
		return a, nil
	})
	for _, version := range []string{LegacySoarVersion, DefaultSoarVersion} {
		current, _ := s.Get(c.ID)
		next, err := s.Edit(c.ID, ClientInput{Revision: current.Revision, Name: c.Name, Enabled: true, SoarVersion: &version})
		if err != nil {
			t.Fatal(err)
		}
		if err = m.Reconcile(); err != nil {
			t.Fatal(err)
		}
		key, _ := s.Seed(c.ID)
		links, _ := s.Profiles(c.ID)
		if !reflect.DeepEqual(seed, key) || !reflect.DeepEqual(profiles, links) || next.PublicKey != c.PublicKey {
			t.Fatal("version changed credentials")
		}
		for _, clientVersion := range []string{"0.2.4", "0.2.3"} {
			client := nativeClient(t, m, next)
			dial := client.DialContext
			if clientVersion == "0.2.3" {
				key, _ := base64.StdEncoding.DecodeString(next.PublicKey)
				old, err := legacy.NewClient(legacy.Config{Zone: next.Zone, ServerKey: ed25519.PublicKey(key), Resolvers: []string{m.pc.LocalAddr().String()}, TCP: "off"})
				if err != nil {
					t.Fatal(err)
				}
				defer old.Close()
				dial = old.DialContext
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			conn, err := dial(ctx, "tcp", "example.com:443")
			if err != nil {
				cancel()
				t.Fatal(err)
			}
			conn.SetDeadline(time.Now().Add(10 * time.Second))
			payload := "version-specific tunnel payload"
			if _, err = io.WriteString(conn, payload); err != nil {
				t.Fatal(err)
			}
			data := make([]byte, len(payload))
			if _, err = io.ReadFull(conn, data); err != nil || string(data) != payload {
				t.Fatalf("transfer: %q %v", data, err)
			}
			conn.Close()
			client.Close()
			cancel()
		}
		saved, err := s.Edit(c.ID, ClientInput{Revision: next.Revision, Name: c.Name, Enabled: true})
		if err != nil || saved.SoarVersion != version {
			t.Fatal("omitted selection lost", err)
		}
	}
	invalid := "latest"
	if _, err := s.Edit(c.ID, ClientInput{Name: c.Name, SoarVersion: &invalid}); err == nil {
		t.Fatal("invalid version accepted")
	}
}

func TestLegacySoarVersionReadAndSnapshot(t *testing.T) {
	s := testStore(t)
	c := account(t, s, 123456)
	raw := clientRecord(t, s, c.ID)
	var record map[string]any
	json.Unmarshal(raw, &record)
	delete(record, "soar_version")
	if err := s.db.Update(func(tx *bolt.Tx) error { return putJSON(tx.Bucket(clients), []byte(c.ID), record) }); err != nil {
		t.Fatal(err)
	}
	before := clientRecord(t, s, c.ID)
	got, err := s.Get(c.ID)
	if err != nil || got.SoarVersion != DefaultSoarVersion || !bytes.Equal(before, clientRecord(t, s, c.ID)) {
		t.Fatal("legacy read rewrote record", err)
	}
	version := LegacySoarVersion
	c, err = s.Edit(c.ID, ClientInput{Revision: got.Revision, Name: got.Name, Enabled: true, SoarVersion: &version})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "backup.db")
	os.WriteFile(path, backupBytes(t, s), 0600)
	if err = ValidateDowngradeSnapshot(path, "0.14.1"); err != nil {
		t.Fatal(err)
	}
	for _, env := range []string{"PANEL_OLDER_PANEL", "PANEL_PREVIOUS_PANEL"} {
		if old := os.Getenv(env); old != "" {
			cmd := exec.Command(old, "inspect-backup", "--input", path)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("older binary rejected snapshot: %v %s", err, out)
			}
		}
	}
	restored, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	saved, err := restored.Get(c.ID)
	if err != nil || saved.SoarVersion != version || saved.PublicKey != c.PublicKey {
		t.Fatal("backup lost version or identity", err)
	}
}
