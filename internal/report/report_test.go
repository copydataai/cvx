package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteRenderReportWritesValidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reports", "last-render.json")
	report := RenderReport{
		Timestamp: "2026-04-30T12:00:00Z",
		Input:     "cv.yaml",
		OutputTeX: "output/cv.tex",
		OutputPDF: "output/cv.pdf",
		Variant:   "variants/backend-engineer.yaml",
		Engine:    "tex-only",
		Validation: ValidationResult{
			Success: true,
		},
		Warnings: []ReportWarning{{
			Code:     "missing_link",
			Message:  "link URL is empty",
			Location: "links[0]",
		}},
		SectionOrder: []string{"summary", "experience"},
	}

	if err := WriteRenderReport(path, report); err != nil {
		t.Fatalf("WriteRenderReport() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if data[len(data)-1] != '\n' {
		t.Fatalf("expected trailing newline, got %q", data)
	}

	var decoded RenderReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if decoded.Input != report.Input || decoded.OutputTeX != report.OutputTeX || decoded.OutputPDF != report.OutputPDF {
		t.Fatalf("decoded output mismatch: %#v", decoded)
	}
	if !decoded.Validation.Success {
		t.Fatalf("expected validation success: %#v", decoded.Validation)
	}
	if len(decoded.Warnings) != 1 || decoded.Warnings[0].Code != "missing_link" {
		t.Fatalf("unexpected warnings: %#v", decoded.Warnings)
	}
	if len(decoded.SectionOrder) != 2 || decoded.SectionOrder[1] != "experience" {
		t.Fatalf("unexpected section order: %#v", decoded.SectionOrder)
	}
}
