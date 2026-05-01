package check

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCommandRunsWorkflowAndSkipsMissingPrevious(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "cv.yaml")
	current := filepath.Join(dir, "output", "current.json")
	renderOutput := filepath.Join(dir, "output", "cv.tex")
	previous := filepath.Join(dir, "output", "previous.json")
	if err := os.WriteFile(input, []byte(`name: Person
summary: Builds tools.
projects:
  - name: Tool
    bullets:
      - Built workflow.
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Command([]string{"--input", input, "--current", current, "--render-output", renderOutput, "--previous", previous}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(current); err != nil {
		t.Fatalf("expected current snapshot: %v", err)
	}
	if _, err := os.Stat(renderOutput); err != nil {
		t.Fatalf("expected render output: %v", err)
	}
}

func TestCommandSavesCurrentSnapshotToHistoryByDefault(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "cv.yaml")
	current := filepath.Join(dir, "output", "current.json")
	renderOutput := filepath.Join(dir, "output", "cv.tex")
	previous := filepath.Join(dir, "output", "previous.json")
	historyDir := filepath.Join(dir, ".cvx", "history")
	if err := os.WriteFile(input, []byte(`name: Person
summary: Builds tools.
projects:
  - name: Tool
    bullets:
      - Built workflow.
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Command([]string{"--input", input, "--current", current, "--render-output", renderOutput, "--previous", previous, "--history-dir", historyDir}); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(historyDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one history snapshot, got %d", len(entries))
	}
	saved := filepath.Join(historyDir, entries[0].Name())
	currentData, err := os.ReadFile(current)
	if err != nil {
		t.Fatal(err)
	}
	savedData, err := os.ReadFile(saved)
	if err != nil {
		t.Fatal(err)
	}
	if string(savedData) != string(currentData) {
		t.Fatalf("history snapshot mismatch")
	}
}
