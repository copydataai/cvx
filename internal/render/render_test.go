package render

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/josesanchez/cvx/internal/cv"
	"github.com/josesanchez/cvx/internal/report"
)

func TestApplyProjectFilterIncludesOnlyNamedProjects(t *testing.T) {
	projects := []cv.Project{{Name: "Keep"}, {Name: "Drop"}}
	variant := &cv.Variant{IncludeProjects: []string{"Keep"}}
	filtered := cv.ApplyVariant(&cv.CV{Projects: projects}, variant).Projects
	if len(filtered) != 1 || filtered[0].Name != "Keep" {
		t.Fatalf("unexpected filtered projects: %#v", filtered)
	}
}

func TestRenderTeXUsesVariantSectionOrder(t *testing.T) {
	doc := &cv.CV{
		Name:    "Person",
		Summary: "Summary",
		Skills:  []string{"Go"},
		Projects: []cv.Project{{
			Name:        "Project",
			Description: "Description",
			Bullets:     []cv.Bullet{{Text: "Built it"}},
		}},
	}
	tex := renderTeX(doc, []string{"skills", "summary", "projects"})
	skillsIndex := strings.Index(tex, "\\section*{Skills}")
	summaryIndex := strings.Index(tex, "\\section*{Summary}")
	if skillsIndex == -1 || summaryIndex == -1 || skillsIndex > summaryIndex {
		t.Fatalf("expected skills before summary, got:\n%s", tex)
	}
}

func TestRenderTeXEscapesLatex(t *testing.T) {
	doc := &cv.CV{Name: "A&B", Summary: "Uses 100%"}
	tex := renderTeX(doc, []string{"summary"})
	if !strings.Contains(tex, "A\\&B") || !strings.Contains(tex, "100\\%") {
		t.Fatalf("expected escaped latex, got:\n%s", tex)
	}
}

func TestCommandWritesRenderReport(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "cv.yaml")
	output := filepath.Join(dir, "out", "cv.tex")
	reportPath := filepath.Join(dir, "reports", "custom-render.json")
	cvYAML := `name: Person
contact:
  email: person@example.com
  location: Dublin
links:
  - label: Website
    url: ""
summary: Builds developer tools.
skills:
  - Go
projects:
  - name: CVX
    description: CV tooling.
    bullets:
      - Built render workflows.
`
	if err := os.WriteFile(input, []byte(cvYAML), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := Command([]string{"--input", input, "--output", output, "--report", reportPath}); err != nil {
		t.Fatalf("Command() error = %v", err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var decoded report.RenderReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Input != input || decoded.OutputTeX != output {
		t.Fatalf("unexpected report paths: %#v", decoded)
	}
	if decoded.OutputPDF != filepath.Join(dir, "out", "cv.pdf") {
		t.Fatalf("unexpected output_pdf: %q", decoded.OutputPDF)
	}
	if decoded.Engine != "tex-only" {
		t.Fatalf("unexpected engine: %q", decoded.Engine)
	}
	if !decoded.Validation.Success || len(decoded.Validation.Errors) != 0 {
		t.Fatalf("unexpected validation result: %#v", decoded.Validation)
	}
	if len(decoded.Warnings) != 1 || decoded.Warnings[0].Code != "missing_link" {
		t.Fatalf("unexpected warnings: %#v", decoded.Warnings)
	}
	if len(decoded.SectionOrder) != 5 || decoded.SectionOrder[0] != "summary" {
		t.Fatalf("unexpected section order: %#v", decoded.SectionOrder)
	}
}

func TestCommandRenderReportWarningsUseFilteredVariantProjects(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "cv.yaml")
	variantPath := filepath.Join(dir, "variant.yaml")
	output := filepath.Join(dir, "out", "cv.tex")
	reportPath := filepath.Join(dir, "reports", "custom-render.json")
	cvYAML := `name: Person
contact:
  email: person@example.com
  location: Dublin
summary: Builds developer tools.
projects:
  - name: Keep
    description: Included project.
    bullets:
      - Built render workflows.
  - name: Drop
    description: Excluded project.
    bullets:
      - ""
`
	variantYAML := `target: backend engineer
max_pages: 1
section_order:
  - summary
  - projects
exclude_projects:
  - Drop
`
	if err := os.WriteFile(input, []byte(cvYAML), 0o644); err != nil {
		t.Fatalf("WriteFile(input) error = %v", err)
	}
	if err := os.WriteFile(variantPath, []byte(variantYAML), 0o644); err != nil {
		t.Fatalf("WriteFile(variant) error = %v", err)
	}

	if err := Command([]string{"--input", input, "--variant", variantPath, "--output", output, "--report", reportPath}); err != nil {
		t.Fatalf("Command() error = %v", err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var decoded report.RenderReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	for _, warning := range decoded.Warnings {
		if warning.Code == "empty_bullet" {
			t.Fatalf("expected excluded project empty bullet warning to be filtered out, got %#v", decoded.Warnings)
		}
	}
}
