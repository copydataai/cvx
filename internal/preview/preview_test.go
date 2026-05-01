package preview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandOnceCreatesHTML(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "cv.yaml")
	output := filepath.Join(dir, "out", "cv.html")
	cvYAML := `name: Person
contact:
  email: person@example.com
  location: Dublin
summary: Builds developer tools.
projects:
  - name: CVX
    description: CV tooling.
    bullets:
      - Built render workflows.
`
	if err := os.WriteFile(input, []byte(cvYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Command([]string{"--once", "--input", input, "--html", output}); err != nil {
		t.Fatalf("Command() error = %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "<html") || !strings.Contains(body, "Person") {
		t.Fatalf("unexpected html output: %s", body)
	}
}
