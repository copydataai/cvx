# Audit Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add agent-auditable warnings, render reports, schema export, and markdown diffs to the Go `cvx` CLI.

**Architecture:** Keep factual CV loading and validation in `internal/cv`, add reusable report writing in `internal/report`, static JSON Schema export in `internal/schema`, and snapshot comparison in `internal/diff`. Wire new commands through `cmd/cvx/main.go` while preserving the current offline, local-first render flow.

**Tech Stack:** Go 1.22, standard library JSON/flag/path/filepath/time packages, `gopkg.in/yaml.v3`, Go tests.

---

## File Structure

- Modify `internal/cv/types.go`: add `Warning` and constants for warning severity/category if needed.
- Modify `internal/cv/load.go`: add `Warnings(doc *CV) []Warning` and helper functions for bullet collection and suspicious metrics.
- Modify `internal/cv/validation_test.go`: test warning collection.
- Create `internal/report/report.go`: define `RenderReport`, `ValidationResult`, and `WriteRenderReport`.
- Create `internal/report/report_test.go`: test JSON report writing and decoding.
- Modify `internal/render/render.go`: collect warnings and write `output/reports/last-render.json` on success and validation failure.
- Modify `internal/render/render_test.go`: test render command report creation through a temp directory.
- Create `internal/schema/schema.go`: static JSON schema writer for CV and variant schemas.
- Create `internal/schema/schema_test.go`: test schema files are written and valid JSON.
- Create `internal/diff/diff.go`: command and markdown generator for CV JSON snapshots.
- Create `internal/diff/diff_test.go`: test added, removed, and changed bullets.
- Modify `cmd/cvx/main.go`: wire `schema` and `diff` commands and update help.

---

### Task 1: Add CV Warning Collection

**Files:**
- Modify: `internal/cv/types.go`
- Modify: `internal/cv/load.go`
- Test: `internal/cv/validation_test.go`

- [ ] **Step 1: Add failing warning tests**

Append these tests to `internal/cv/validation_test.go`:

```go
func TestWarningsDetectAgentRelevantIssues(t *testing.T) {
	doc := &CV{
		Name:    "Person",
		Summary: "Summary",
		Links: []Link{
			{Label: "GitHub", URL: ""},
		},
		Skills: []string{"Go", "Linux", "Postgres", "Redis", "Docker", "Kubernetes", "AWS", "Terraform", "Python", "TypeScript", "React", "GraphQL", "Prometheus", "NATS", "SQLite", "Bash"},
		Projects: []Project{
			{Name: "P1", Bullets: []string{"", "Built workflow", "Built workflow", "Improved conversion by 40%"}},
			{Name: "P2"}, {Name: "P3"}, {Name: "P4"}, {Name: "P5"}, {Name: "P6"}, {Name: "P7"},
		},
	}

	warnings := Warnings(doc)
	codes := map[string]bool{}
	for _, warning := range warnings {
		codes[warning.Code] = true
	}

	for _, code := range []string{"missing_link", "too_many_skills", "too_many_projects", "empty_bullet", "duplicate_bullet", "suspicious_metric"} {
		if !codes[code] {
			t.Fatalf("expected warning code %q in %#v", code, warnings)
		}
	}
}

func TestWarningsDetectLongBullets(t *testing.T) {
	long := strings.Repeat("word ", 46)
	doc := &CV{Name: "Person", Summary: "Summary", Projects: []Project{{Name: "P", Bullets: []string{long}}}}
	warnings := Warnings(doc)
	for _, warning := range warnings {
		if warning.Code == "long_bullet" {
			return
		}
	}
	t.Fatalf("expected long_bullet warning in %#v", warnings)
}
```

Also add `strings` to the test import block:

```go
import (
	"strings"
	"testing"
)
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/cv
```

Expected: fail with `undefined: Warnings` or `undefined: Warning`.

- [ ] **Step 3: Add warning model and implementation**

Add to `internal/cv/types.go`:

```go
type Warning struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Location string `json:"location"`
}
```

Add to `internal/cv/load.go`:

```go
func Warnings(doc *CV) []Warning {
	var warnings []Warning
	if len(doc.Skills) > 15 {
		warnings = append(warnings, Warning{Code: "too_many_skills", Message: "skills list has more than 15 entries", Location: "skills"})
	}
	if len(doc.Projects) > 6 {
		warnings = append(warnings, Warning{Code: "too_many_projects", Message: "projects list has more than 6 entries", Location: "projects"})
	}
	for i, link := range doc.Links {
		if strings.TrimSpace(link.URL) == "" {
			warnings = append(warnings, Warning{Code: "missing_link", Message: "link is missing a URL", Location: fmt.Sprintf("links[%d]", i)})
		}
	}
	seenBullets := map[string]string{}
	for _, bullet := range collectBullets(doc) {
		text := strings.TrimSpace(bullet.Text)
		if text == "" {
			warnings = append(warnings, Warning{Code: "empty_bullet", Message: "bullet is empty", Location: bullet.Location})
			continue
		}
		if wordCount(text) > 45 {
			warnings = append(warnings, Warning{Code: "long_bullet", Message: "bullet is longer than 45 words", Location: bullet.Location})
		}
		if first, ok := seenBullets[text]; ok {
			warnings = append(warnings, Warning{Code: "duplicate_bullet", Message: "bullet duplicates " + first, Location: bullet.Location})
		} else {
			seenBullets[text] = bullet.Location
		}
		if hasSuspiciousMetric(text) {
			warnings = append(warnings, Warning{Code: "suspicious_metric", Message: "bullet contains a metric; confirm it is sourced", Location: bullet.Location})
		}
	}
	return warnings
}

type bulletRef struct {
	Location string
	Text     string
}

func collectBullets(doc *CV) []bulletRef {
	var bullets []bulletRef
	for i, exp := range doc.Experience {
		for j, bullet := range exp.Bullets {
			bullets = append(bullets, bulletRef{Location: fmt.Sprintf("experience[%d].bullets[%d]", i, j), Text: bullet})
		}
	}
	for i, project := range doc.Projects {
		for j, bullet := range project.Bullets {
			bullets = append(bullets, bulletRef{Location: fmt.Sprintf("projects[%d].bullets[%d]", i, j), Text: bullet})
		}
	}
	return bullets
}

func wordCount(text string) int {
	return len(strings.Fields(text))
}

func hasSuspiciousMetric(text string) bool {
	if strings.Contains(text, "[Sourced]") || strings.Contains(text, "[Verified]") {
		return false
	}
	for _, token := range strings.Fields(text) {
		if strings.Contains(token, "%") {
			return true
		}
		for _, r := range token {
			if r >= '0' && r <= '9' {
				return true
			}
		}
	}
	return false
}
```

`internal/cv/load.go` already imports `fmt` and `strings`; keep those imports.

- [ ] **Step 4: Run tests**

Run:

```bash
gofmt -w internal/cv/types.go internal/cv/load.go internal/cv/validation_test.go
go test ./internal/cv
```

Expected: pass.

- [ ] **Step 5: Commit**

```bash
git add internal/cv/types.go internal/cv/load.go internal/cv/validation_test.go
git commit -m "feat: add cv warning collection"
```

---

### Task 2: Add Render Report Writer and Wire Render

**Files:**
- Create: `internal/report/report.go`
- Create: `internal/report/report_test.go`
- Modify: `internal/render/render.go`
- Modify: `internal/render/render_test.go`

- [ ] **Step 1: Add report writer test**

Create `internal/report/report_test.go`:

```go
package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteRenderReport(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "last-render.json")
	report := RenderReport{
		Timestamp: "2026-04-30T12:00:00Z",
		Input:     "cv.yaml",
		OutputTeX: "output/cv.tex",
		OutputPDF: "output/cv.pdf",
		Engine:    "tex-only",
		Validation: ValidationResult{
			Success: true,
		},
		Warnings: []ReportWarning{{Code: "missing_link", Message: "link is missing a URL", Location: "links[0]"}},
		SectionOrder: []string{"summary", "projects"},
	}

	if err := WriteRenderReport(path, report); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var decoded RenderReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Validation.Success || decoded.Engine != "tex-only" || len(decoded.Warnings) != 1 {
		t.Fatalf("unexpected report: %#v", decoded)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/report
```

Expected: fail because package files/types do not exist.

- [ ] **Step 3: Implement report writer**

Create `internal/report/report.go`:

```go
package report

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type RenderReport struct {
	Timestamp    string             `json:"timestamp"`
	Input        string             `json:"input"`
	OutputTeX    string             `json:"output_tex"`
	OutputPDF    string             `json:"output_pdf"`
	Template     string             `json:"template,omitempty"`
	Variant      string             `json:"variant,omitempty"`
	Engine       string             `json:"engine"`
	Validation   ValidationResult   `json:"validation"`
	Warnings     []ReportWarning    `json:"warnings"`
	SectionOrder []string           `json:"section_order"`
}

type ValidationResult struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}

type ReportWarning struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Location string `json:"location"`
}

func WriteRenderReport(path string, report RenderReport) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
```

- [ ] **Step 4: Wire render report into render command**

In `internal/render/render.go`, add imports:

```go
	"time"

	"github.com/josesanchez/cvx/internal/report"
```

Add flag in `Command`:

```go
	reportPath := fs.String("report", "output/reports/last-render.json", "render report JSON path")
```

After loading `doc`, compute defaults:

```go
	sections := []string{"summary", "experience", "projects", "skills", "education"}
	warnings := cv.Warnings(doc)
```

Replace the existing validation failure return with:

```go
	if errs := cv.Validate(doc); len(errs) > 0 {
		renderReport := newRenderReport(*input, *output, *variantPath, sections, warnings, false, errs)
		if err := report.WriteRenderReport(*reportPath, renderReport); err != nil {
			return err
		}
		return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
	}
```

After variant project filtering, recompute warnings for rendered content:

```go
	warnings = cv.Warnings(doc)
```

After writing TeX, write success report:

```go
	renderReport := newRenderReport(*input, *output, *variantPath, sections, warnings, true, nil)
	if err := report.WriteRenderReport(*reportPath, renderReport); err != nil {
		return err
	}
```

Add helper functions to `internal/render/render.go`:

```go
func newRenderReport(input, output, variant string, sections []string, warnings []cv.Warning, success bool, errors []string) report.RenderReport {
	return report.RenderReport{
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Input:        input,
		OutputTeX:    output,
		OutputPDF:    defaultPDFPath(output),
		Variant:      variant,
		Engine:       "tex-only",
		Validation:   report.ValidationResult{Success: success, Errors: errors},
		Warnings:     reportWarnings(warnings),
		SectionOrder: sections,
	}
}

func defaultPDFPath(texPath string) string {
	if strings.HasSuffix(texPath, ".tex") {
		return strings.TrimSuffix(texPath, ".tex") + ".pdf"
	}
	return texPath + ".pdf"
}

func reportWarnings(warnings []cv.Warning) []report.ReportWarning {
	converted := make([]report.ReportWarning, 0, len(warnings))
	for _, warning := range warnings {
		converted = append(converted, report.ReportWarning{Code: warning.Code, Message: warning.Message, Location: warning.Location})
	}
	return converted
}
```

- [ ] **Step 5: Add render command report test**

Append to `internal/render/render_test.go`:

```go
func TestCommandWritesRenderReport(t *testing.T) {
	dir := t.TempDir()
	cvPath := filepath.Join(dir, "cv.yaml")
	texPath := filepath.Join(dir, "output", "cv.tex")
	reportPath := filepath.Join(dir, "output", "reports", "last-render.json")
	body := `name: Person
contact:
  email: person@example.com
  location: Remote
summary: Summary
skills:
  - Go
projects:
  - name: Project
    description: Description
    bullets:
      - Built project
`
	if err := os.WriteFile(cvPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Command([]string{"--input", cvPath, "--output", texPath, "--report", reportPath}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"engine": "tex-only"`) || !strings.Contains(string(data), `"success": true`) {
		t.Fatalf("unexpected report: %s", string(data))
	}
}
```

Add imports to `internal/render/render_test.go`:

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)
```

- [ ] **Step 6: Run tests**

Run:

```bash
gofmt -w internal/report internal/render
go test ./...
```

Expected: pass.

- [ ] **Step 7: Commit**

```bash
git add internal/report/report.go internal/report/report_test.go internal/render/render.go internal/render/render_test.go
git commit -m "feat: write render audit reports"
```

---

### Task 3: Add JSON Schema Export

**Files:**
- Create: `internal/schema/schema.go`
- Create: `internal/schema/schema_test.go`
- Modify: `cmd/cvx/main.go`

- [ ] **Step 1: Add schema tests**

Create `internal/schema/schema_test.go`:

```go
package schema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteSchemas(t *testing.T) {
	dir := t.TempDir()
	if err := WriteSchemas(dir); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"cv.schema.json", "variant.schema.json"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("%s is not valid JSON: %v", name, err)
		}
		if decoded["$schema"] == "" || decoded["type"] != "object" {
			t.Fatalf("%s has unexpected schema shape: %#v", name, decoded)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/schema
```

Expected: fail because `WriteSchemas` does not exist.

- [ ] **Step 3: Implement schema writer**

Create `internal/schema/schema.go` with static schema constants and writer:

```go
package schema

import (
	"os"
	"path/filepath"
)

const cvSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/cvx/cv.schema.json",
  "type": "object",
  "required": ["name", "summary"],
  "properties": {
    "name": {"type": "string"},
    "contact": {
      "type": "object",
      "properties": {
        "email": {"type": "string"},
        "phone": {"type": "string"},
        "location": {"type": "string"}
      }
    },
    "links": {"type": "array", "items": {"type": "object", "properties": {"label": {"type": "string"}, "url": {"type": "string"}}}},
    "summary": {"type": "string"},
    "skills": {"type": "array", "items": {"type": "string"}},
    "experience": {"type": "array", "items": {"type": "object", "properties": {"company": {"type": "string"}, "title": {"type": "string"}, "location": {"type": "string"}, "start": {"type": "string"}, "end": {"type": "string"}, "bullets": {"type": "array", "items": {"type": "string"}}}}},
    "projects": {"type": "array", "items": {"type": "object", "properties": {"name": {"type": "string"}, "description": {"type": "string"}, "url": {"type": "string"}, "bullets": {"type": "array", "items": {"type": "string"}}}}},
    "education": {"type": "array", "items": {"type": "object", "properties": {"institution": {"type": "string"}, "degree": {"type": "string"}, "start": {"type": "string"}, "end": {"type": "string"}}}},
    "metadata": {"type": "object", "properties": {"updated": {"type": "string"}}}
  }
}
`

const variantSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/cvx/variant.schema.json",
  "type": "object",
  "required": ["target", "section_order"],
  "properties": {
    "target": {"type": "string"},
    "max_pages": {"type": "integer", "minimum": 0},
    "tone": {"type": "string"},
    "section_order": {"type": "array", "items": {"type": "string", "enum": ["summary", "experience", "projects", "skills", "education"]}},
    "include_projects": {"type": "array", "items": {"type": "string"}},
    "exclude_projects": {"type": "array", "items": {"type": "string"}},
    "emphasis_keywords": {"type": "array", "items": {"type": "string"}}
  }
}
`

func WriteSchemas(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "cv.schema.json"), []byte(cvSchema), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "variant.schema.json"), []byte(variantSchema), 0o644)
}
```

- [ ] **Step 4: Wire CLI command**

In `cmd/cvx/main.go`, add import:

```go
	"github.com/josesanchez/cvx/internal/schema"
```

Add switch case:

```go
	case "schema":
		if err := schema.WriteSchemas("schema"); err != nil {
			return err
		}
		fmt.Println("wrote schema/cv.schema.json")
		fmt.Println("wrote schema/variant.schema.json")
		return nil
```

Update usage text to include:

```txt
  cvx schema
```

- [ ] **Step 5: Run tests and command**

Run:

```bash
gofmt -w cmd/cvx/main.go internal/schema
go test ./...
go run ./cmd/cvx schema
```

Expected: tests pass and schema files are written.

- [ ] **Step 6: Commit**

```bash
git add cmd/cvx/main.go internal/schema/schema.go internal/schema/schema_test.go schema/cv.schema.json schema/variant.schema.json
git commit -m "feat: export json schemas"
```

---

### Task 4: Add Markdown Diff Command

**Files:**
- Create: `internal/diff/diff.go`
- Create: `internal/diff/diff_test.go`
- Modify: `cmd/cvx/main.go`

- [ ] **Step 1: Add diff tests**

Create `internal/diff/diff_test.go`:

```go
package diff

import (
	"strings"
	"testing"

	"github.com/josesanchez/cvx/internal/cv"
)

func TestMarkdownReportsAddedRemovedAndChangedBullets(t *testing.T) {
	from := &cv.CV{
		Summary: "Old summary",
		Projects: []cv.Project{{Name: "P", Bullets: []string{"Old bullet", "Removed bullet"}}},
	}
	to := &cv.CV{
		Summary: "New summary",
		Projects: []cv.Project{{Name: "P", Bullets: []string{"New bullet", "Added bullet"}}},
	}
	report := Markdown(from, to)
	for _, expected := range []string{"summary", "Changed wording", "Old bullet", "New bullet", "Removed bullet", "Added bullet"} {
		if !strings.Contains(report, expected) {
			t.Fatalf("expected %q in report:\n%s", expected, report)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/diff
```

Expected: fail because `Markdown` does not exist.

- [ ] **Step 3: Implement diff package**

Create `internal/diff/diff.go`:

```go
package diff

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/josesanchez/cvx/internal/cv"
)

func Command(args []string) error {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	fromPath := fs.String("from", "", "previous CV JSON snapshot")
	toPath := fs.String("to", "", "current CV JSON snapshot")
	output := fs.String("output", "output/reports/last-diff.md", "diff markdown report path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *fromPath == "" || *toPath == "" {
		return fmt.Errorf("diff requires --from and --to")
	}
	from, err := loadJSON(*fromPath)
	if err != nil {
		return fmt.Errorf("load --from %s: %w", *fromPath, err)
	}
	to, err := loadJSON(*toPath)
	if err != nil {
		return fmt.Errorf("load --to %s: %w", *toPath, err)
	}
	body := Markdown(from, to)
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(*output, []byte(body), 0o644); err != nil {
		return err
	}
	fmt.Println("wrote", *output)
	return nil
}

func loadJSON(path string) (*cv.CV, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc cv.CV
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func Markdown(from, to *cv.CV) string {
	var b strings.Builder
	b.WriteString("# CV Diff Report\n\n")
	changedSections := changedSections(from, to)
	if len(changedSections) == 0 {
		b.WriteString("No section-level changes detected.\n\n")
	} else {
		b.WriteString("## Changed Sections\n\n")
		for _, section := range changedSections {
			b.WriteString("- " + section + "\n")
		}
		b.WriteString("\n")
	}
	writeBulletDiff(&b, "Experience", experienceBullets(from), experienceBullets(to))
	writeBulletDiff(&b, "Projects", projectBullets(from), projectBullets(to))
	return b.String()
}

func changedSections(from, to *cv.CV) []string {
	var sections []string
	if from.Summary != to.Summary {
		sections = append(sections, "summary")
	}
	if strings.Join(from.Skills, "\x00") != strings.Join(to.Skills, "\x00") {
		sections = append(sections, "skills")
	}
	if len(from.Experience) != len(to.Experience) {
		sections = append(sections, "experience")
	}
	if len(from.Projects) != len(to.Projects) {
		sections = append(sections, "projects")
	}
	if len(from.Education) != len(to.Education) {
		sections = append(sections, "education")
	}
	return sections
}

func writeBulletDiff(b *strings.Builder, title string, from, to []string) {
	b.WriteString("## " + title + " Bullets\n\n")
	max := len(from)
	if len(to) > max {
		max = len(to)
	}
	wrote := false
	for i := 0; i < max; i++ {
		switch {
		case i >= len(from):
			b.WriteString("- Added bullet: " + to[i] + "\n")
			wrote = true
		case i >= len(to):
			b.WriteString("- Removed bullet: " + from[i] + "\n")
			wrote = true
		case from[i] != to[i]:
			b.WriteString("- Changed wording:\n")
			b.WriteString("  - From: " + from[i] + "\n")
			b.WriteString("  - To: " + to[i] + "\n")
			wrote = true
		}
	}
	if !wrote {
		b.WriteString("No bullet changes detected.\n")
	}
	b.WriteString("\n")
}

func experienceBullets(doc *cv.CV) []string {
	var bullets []string
	for _, exp := range doc.Experience {
		bullets = append(bullets, exp.Bullets...)
	}
	return bullets
}

func projectBullets(doc *cv.CV) []string {
	var bullets []string
	for _, project := range doc.Projects {
		bullets = append(bullets, project.Bullets...)
	}
	return bullets
}
```

- [ ] **Step 4: Wire CLI command**

In `cmd/cvx/main.go`, add import:

```go
	"github.com/josesanchez/cvx/internal/diff"
```

Add switch case:

```go
	case "diff":
		return diff.Command(args[1:])
```

Update usage text to include:

```txt
  cvx diff --from output/previous.json --to output/current.json
```

- [ ] **Step 5: Run tests**

Run:

```bash
gofmt -w cmd/cvx/main.go internal/diff
go test ./...
```

Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add cmd/cvx/main.go internal/diff/diff.go internal/diff/diff_test.go
git commit -m "feat: add cv diff reports"
```

---

### Task 5: Final Verification and Documentation Touch

**Files:**
- Modify: `AGENTS.md` if workflow commands need the new audit commands.
- Modify: `docs/superpowers/specs/2026-04-30-audit-foundation-design.md` only if implementation discovers a mismatch that must be recorded.

- [ ] **Step 1: Run full verification**

Run:

```bash
go test ./...
go run ./cmd/cvx lint --variant variants/yc-founder-engineer.yaml
go run ./cmd/cvx render --variant variants/yc-founder-engineer.yaml
go run ./cmd/cvx schema
```

Expected: all commands pass.

- [ ] **Step 2: Inspect generated audit artifacts**

Run:

```bash
cat output/reports/last-render.json
ls schema/cv.schema.json schema/variant.schema.json
```

Expected: render report exists with `engine` set to `tex-only`; schema files exist.

- [ ] **Step 3: Commit documentation if changed**

If `AGENTS.md` or docs changed, run:

```bash
git add AGENTS.md docs/superpowers/specs/2026-04-30-audit-foundation-design.md
git commit -m "docs: document audit commands"
```

If no docs changed, do not create an empty commit.

- [ ] **Step 4: Final status**

Run:

```bash
git status --short --branch
```

Expected: only ignored/generated files are untracked or the working tree is clean.
