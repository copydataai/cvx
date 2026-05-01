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
