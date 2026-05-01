package report

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type RenderReport struct {
	Timestamp    string           `json:"timestamp"`
	Input        string           `json:"input"`
	OutputTeX    string           `json:"output_tex,omitempty"`
	OutputHTML   string           `json:"output_html,omitempty"`
	OutputPDF    string           `json:"output_pdf,omitempty"`
	Template     string           `json:"template,omitempty"`
	Variant      string           `json:"variant,omitempty"`
	Engine       string           `json:"engine"`
	Validation   ValidationResult `json:"validation"`
	Warnings     []ReportWarning  `json:"warnings"`
	SectionOrder []string         `json:"section_order"`
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
