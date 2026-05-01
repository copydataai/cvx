package history

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestSaveSnapshotCopiesCurrentSnapshotToTimestampedHistoryFile(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "output", "current.json")
	historyDir := filepath.Join(dir, ".cvx", "history")
	data := []byte("{\n  \"name\": \"Person\"\n}\n")

	if err := os.MkdirAll(filepath.Dir(current), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(current, data, 0o644); err != nil {
		t.Fatal(err)
	}

	saved, err := SaveSnapshot(current, historyDir)
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Dir(saved) != historyDir {
		t.Fatalf("expected history dir %s, got %s", historyDir, filepath.Dir(saved))
	}
	if !regexp.MustCompile(`^\d{8}T\d{6}Z\.json$`).MatchString(filepath.Base(saved)) {
		t.Fatalf("expected UTC timestamp filename, got %s", filepath.Base(saved))
	}
	got, err := os.ReadFile(saved)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Fatalf("saved snapshot mismatch: got %q want %q", got, data)
	}
}
