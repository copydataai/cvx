package normalize

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/josesanchez/cvx/internal/cv"
)

func TestWriteAppliesVariantProjectFiltering(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "cv.yaml")
	variant := filepath.Join(dir, "variant.yaml")
	output := filepath.Join(dir, "output", "current.json")

	cvYAML := `name: Person
summary: Builds tools.
projects:
  - name: Keep
    description: Kept project.
    bullets:
      - Built workflow.
  - name: Drop
    description: Dropped project.
    bullets:
      - Should not appear.
`
	variantYAML := `target: backend engineer
section_order:
  - summary
  - projects
exclude_projects:
  - Drop
`
	if err := os.WriteFile(input, []byte(cvYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(variant, []byte(variantYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Write(input, variant, output); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var doc cv.CV
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Projects) != 1 || doc.Projects[0].Name != "Keep" {
		t.Fatalf("unexpected projects: %#v", doc.Projects)
	}
}

func TestWriteFailsValidation(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "cv.yaml")
	output := filepath.Join(dir, "output", "current.json")
	if err := os.WriteFile(input, []byte("summary: Missing name\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Write(input, "", output); err == nil {
		t.Fatal("expected validation error")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("expected no output file, stat err = %v", err)
	}
}
