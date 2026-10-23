package panel

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"strings"

	bolt "go.etcd.io/bbolt"
)

func (s Settings) accountLabelLength() int {
	if s.AccountLabelLength == 0 {
		return DefaultAccountLabelLength
	}
	return s.AccountLabelLength
}

func randomAccountLabel(length int) (string, error) {
	raw := make([]byte, (length*5+7)/8)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw))[:length], nil
}

// Called only inside the write transaction that creates or rotates an account.
// A maximum-length base zone leaves 20 label characters; reserving that prefix
// also prevents collisions when alternate tunnel domains truncate long labels.
func allocateAccountZone(tx *bolt.Tx, cfg Settings, draw func(int) (string, error)) (string, error) {
	used := map[string]bool{}
	if err := tx.Bucket(clients).ForEach(func(_, raw []byte) error {
		var c Client
		if err := json.Unmarshal(raw, &c); err != nil {
			return err
		}
		label, _, _ := strings.Cut(c.Zone, ".")
		used[label[:min(20, len(label))]] = true
		return nil
	}); err != nil {
		return "", err
	}
	length := min(cfg.accountLabelLength(), 44-len(cfg.BaseZone))
	if length < 3 || length > 26 {
		return "", errors.New("account_label_length must be between 3 and 26")
	}
	for attempt := 0; attempt < 256; attempt++ {
		label, err := draw(length)
		if err != nil {
			return "", err
		}
		if !used[label[:min(20, len(label))]] {
			return label + "." + cfg.BaseZone, nil
		}
	}
	return "", errors.New("cannot allocate a unique account domain; retry or increase account label length")
}
