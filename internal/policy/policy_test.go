package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPolicyBlocksNameChange(t *testing.T) {
	dir := t.TempDir()
	before := filepath.Join(dir, "before.json")
	after := filepath.Join(dir, "after.json")
	if err := os.WriteFile(before, []byte(`{"cv":{"name":"Before","summary":"S","projects":[{"name":"P"}]}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(after, []byte(`{"cv":{"name":"After","summary":"S","projects":[{"name":"P"}]}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := Check(before, after, false)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "blocked" || report.Summary.High != 1 {
		t.Fatalf("unexpected report: %#v", report)
	}
}
