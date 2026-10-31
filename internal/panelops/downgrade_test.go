package panelops

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"soar-server-setup/internal/panel"
)

func TestDowngradeActivationKeepsMaintenanceToolsAndRollsBack(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX fixture")
	}
	for _, scenario := range []string{"success", "startup-failure", "unreadable-database", "bad-download", "newer-target"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			db := filepath.Join(dir, "panel.db")
			database(t, db, "unchanged-password-2026")
			beforeDB, _ := os.ReadFile(db)
			binary := filepath.Join(dir, "panel")
			ops := filepath.Join(dir, "ops")
			security := filepath.Join(dir, "security")
			menu := filepath.Join(dir, "menu")
			original := []byte("#!/bin/sh\nprintf '0.14.2\\n'\n")
			os.WriteFile(binary, original, 0755)
			for _, p := range []string{ops, security, menu} {
				os.WriteFile(p, []byte("current-maintenance-tool"), 0755)
			}
			stops, starts := 0, 0
			target := "0.14.1"
			if scenario == "newer-target" {
				target = "0.14.3"
			}
			s := &Service{DB: db, Binary: binary, OpsBinary: ops, SecurityScript: security, MenuScript: menu, StateDir: filepath.Join(dir, "jobs"), UID: -1, GID: -1,
				Control: func(_ context.Context, action string) error {
					if action == "stop" {
						stops++
					}
					if action == "start" {
						starts++
						if scenario == "startup-failure" && starts == 1 {
							return errors.New("startup failed")
						}
					}
					return nil
				}, Health: func(context.Context) error { return nil },
				Fetch: func(_ context.Context, v, work string) error {
					if scenario == "bad-download" {
						return errors.New("checksum mismatch")
					}
					script := "#!/bin/sh\nif [ \"$1\" = version ]; then printf '" + v + "\\n'; else printf '{}\\n'; fi\n"
					if scenario == "unreadable-database" {
						script = "#!/bin/sh\nif [ \"$1\" = version ]; then printf '" + v + "\\n'; else exit 1; fi\n"
					}
					for _, name := range []string{"soar-panel", "soar-panel-ops"} {
						if err := os.WriteFile(filepath.Join(work, name), []byte(script), 0755); err != nil {
							return err
						}
					}
					return nil
				}}
			if err := s.Initialize(); err != nil {
				t.Fatal(err)
			}
			s.perform(panel.MaintenanceRequest{Action: "downgrade", Version: target}, panel.MaintenanceJob{ID: "downgrade-fixture", Kind: "downgrade", State: "running"})
			job := s.Status()
			switch scenario {
			case "success":
				if job.State != "succeeded" {
					t.Fatal(job)
				}
			case "bad-download", "newer-target":
				if stops != 0 || job.State != "failed" {
					t.Fatal(job, stops)
				}
			default:
				if job.State != "rolled_back" {
					t.Fatal(job)
				}
			}
			for _, p := range []string{ops, security, menu} {
				data, _ := os.ReadFile(p)
				if string(data) != "current-maintenance-tool" {
					t.Fatal("maintenance tool replaced", p)
				}
			}
			data, _ := os.ReadFile(db)
			if !bytes.Equal(data, beforeDB) {
				t.Fatal("database changed")
			}
			if scenario != "success" {
				data, _ = os.ReadFile(binary)
				if !bytes.Equal(data, original) {
					t.Fatal("panel not restored")
				}
			}
		})
	}
}
