package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFactsReportGeneration(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "cv.yaml")
	output := filepath.Join(dir, "reports", "review-facts.md")
	writeFile(t, input, `name: Ada Lovelace
contact:
  email: ada@example.com
links:
  - label: Portfolio
    url: ""
summary: Systems builder.
skills:
  - Go
experience:
  - company: Analytical Engines
    title: Founder Engineer
    start: "2024"
    end: Present
    bullets:
      - Increased conversion by 20%.
projects: []
education: []
`)

	if err := Command([]string{"facts", "--input", input, "--output", output}); err != nil {
		t.Fatalf("Command() error = %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(data)
	assertContains(t, text, "# CV Ada Lovelace Review")
	assertContains(t, text, "## Facts")
	assertContains(t, text, "- missing_link at links[0]: link URL is empty")
	assertContains(t, text, "- suspicious_metric at experience[0].bullets[0]: bullet contains a metric without [Sourced] or [Verified]")
}

func TestTargetReportGeneration(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "cv.yaml")
	variant := filepath.Join(dir, "variant.yaml")
	output := filepath.Join(dir, "reports", "review-target.md")
	writeFile(t, input, `name: Ada Lovelace
contact:
  email: ada@example.com
links: []
summary: Systems builder.
skills:
  - Go
experience: []
projects:
  - name: cvx
    description: Agent-native CV system.
    bullets:
      - Built deterministic reviews.
education: []
`)
	writeFile(t, variant, `target: YC founder engineer
section_order:
  - summary
  - projects
  - experience
emphasis_keywords:
  - systems
  - product
include_projects:
  - cvx
  - missing-project
exclude_projects:
  - old-project
`)

	if err := Command([]string{"target", "--input", input, "--variant", variant, "--output", output}); err != nil {
		t.Fatalf("Command() error = %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(data)
	assertContains(t, text, "# CV Ada Lovelace Review")
	assertContains(t, text, "## Target")
	assertContains(t, text, "### Target\n- YC founder engineer")
	assertContains(t, text, "### Section Order\n- summary\n- projects\n- experience")
	assertContains(t, text, "### Emphasis Keywords\n- systems\n- product")
	assertContains(t, text, "### Included Projects\n- cvx\n- missing-project")
	assertContains(t, text, "### Excluded Projects\n- old-project")
	assertContains(t, text, "### Missing Included Projects\n- missing-project")
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func assertContains(t *testing.T, text, want string) {
	t.Helper()
	if !strings.Contains(text, want) {
		t.Fatalf("expected report to contain %q\nreport:\n%s", want, text)
	}
}
