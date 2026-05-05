package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStartAndSnapshot(t *testing.T) {
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("cv.yaml", []byte("name: Person\nsummary: Summary\nprojects:\n  - name: P\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sessionDir := filepath.Join(dir, ".cvx", "sessions", "test")
	if err := start([]string{"--goal", "Test goal", "--dir", sessionDir}); err != nil {
		t.Fatal(err)
	}
	if err := snapshot([]string{"--dir", sessionDir, "--label", "before"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(sessionDir, "before.json")); err != nil {
		t.Fatal(err)
	}
}
