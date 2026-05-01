package render

import (
	"flag"
	"fmt"
	htmlstd "html"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/josesanchez/cvx/internal/cv"
	"github.com/josesanchez/cvx/internal/report"
)

func Command(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	input := fs.String("input", "cv.yaml", "CV YAML path")
	variantPath := fs.String("variant", "", "variant YAML path")
	output := fs.String("output", "", "render output path")
	format := fs.String("format", "tex", "render format: tex or html")
	templatePath := fs.String("template", "", "template path for html format")
	reportPath := fs.String("report", "output/reports/last-render.json", "render report JSON path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	sections := []string{"summary", "experience", "projects", "skills", "education"}
	doc, err := cv.Load(*input)
	if err != nil {
		return err
	}
	actualOutput := defaultOutputPath(*format, *output)
	renderReport := newRenderReport(*input, actualOutput, *format, *variantPath, sections, nil)
	if errs := cv.Validate(doc); len(errs) > 0 {
		renderReport.Warnings = reportWarnings(cv.Warnings(doc))
		renderReport.Validation = report.ValidationResult{Success: false, Errors: errs}
		if err := report.WriteRenderReport(*reportPath, renderReport); err != nil {
			return fmt.Errorf("write render report: %w", err)
		}
		return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
	}

	if *variantPath != "" {
		variant, err := cv.LoadVariant(*variantPath)
		if err != nil {
			return err
		}
		if errs := cv.ValidateVariant(variant); len(errs) > 0 {
			renderReport.SectionOrder = append([]string(nil), variant.SectionOrder...)
			renderReport.Warnings = reportWarnings(cv.Warnings(doc))
			renderReport.Validation = report.ValidationResult{Success: false, Errors: errs}
			if err := report.WriteRenderReport(*reportPath, renderReport); err != nil {
				return fmt.Errorf("write render report: %w", err)
			}
			return fmt.Errorf("variant validation failed: %s", strings.Join(errs, "; "))
		}
		sections = variant.SectionOrder
		renderReport.SectionOrder = append([]string(nil), sections...)
		doc = cv.ApplyVariant(doc, variant)
	}

	body, err := renderBody(doc, sections, *format, *templatePath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(actualOutput), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(actualOutput, []byte(body), 0o644); err != nil {
		return err
	}
	renderReport.Warnings = reportWarnings(cv.Warnings(doc))
	renderReport.Validation = report.ValidationResult{Success: true}
	if err := report.WriteRenderReport(*reportPath, renderReport); err != nil {
		return err
	}
	fmt.Println("rendered", actualOutput)
	return nil
}

func newRenderReport(input, output, format, variantPath string, sections []string, warnings []report.ReportWarning) report.RenderReport {
	report := report.RenderReport{
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Input:        input,
		Variant:      variantPath,
		Engine:       engineName(format),
		Warnings:     warnings,
		SectionOrder: append([]string(nil), sections...),
	}
	if format == "html" {
		report.OutputHTML = output
		return report
	}
	report.OutputTeX = output
	report.OutputPDF = outputPDFPath(output)
	return report
}

func engineName(format string) string {
	if format == "html" {
		return "html"
	}
	return "tex-only"
}

func outputPDFPath(outputTeX string) string {
	if strings.HasSuffix(outputTeX, ".tex") {
		return strings.TrimSuffix(outputTeX, ".tex") + ".pdf"
	}
	return outputTeX + ".pdf"
}

func reportWarnings(warnings []cv.Warning) []report.ReportWarning {
	result := make([]report.ReportWarning, 0, len(warnings))
	for _, warning := range warnings {
		result = append(result, report.ReportWarning{
			Code:     warning.Code,
			Message:  warning.Message,
			Location: warning.Location,
		})
	}
	return result
}

func defaultOutputPath(format, output string) string {
	if output != "" {
		return output
	}
	if format == "html" {
		return "output/cv.html"
	}
	return "output/cv.tex"
}

func renderBody(doc *cv.CV, sections []string, format, templatePath string) (string, error) {
	switch format {
	case "", "tex":
		return renderTeX(doc, sections), nil
	case "html":
		if templatePath == "" {
			templatePath = defaultHTMLTemplatePath()
		}
		return renderHTML(doc, templatePath)
	default:
		return "", fmt.Errorf("unknown render format %q", format)
	}
}

func defaultHTMLTemplatePath() string {
	for _, path := range []string{
		"templates/html/minimal.html.tmpl",
		"../../templates/html/minimal.html.tmpl",
	} {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return "templates/html/minimal.html.tmpl"
}

func renderHTML(doc *cv.CV, templatePath string) (string, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("parse html template %s: %w", templatePath, err)
	}
	var b strings.Builder
	if err := tmpl.Execute(&b, doc); err != nil {
		return "", fmt.Errorf("execute html template %s: %w", templatePath, err)
	}
	return b.String(), nil
}

func renderTeX(doc *cv.CV, sections []string) string {
	var b strings.Builder
	b.WriteString("\\documentclass[11pt]{article}\n\\usepackage[margin=0.7in]{geometry}\n\\begin{document}\n")
	b.WriteString("\\begin{center}{\\LARGE " + tex(doc.Name) + "}\\\\\n")
	b.WriteString(tex(doc.Contact.Email) + " " + tex(doc.Contact.Location) + "\\end{center}\n")
	for _, section := range sections {
		switch section {
		case "summary":
			if doc.Summary != "" {
				b.WriteString("\\section*{Summary}\n" + tex(doc.Summary) + "\n")
			}
		case "skills":
			if len(doc.Skills) > 0 {
				b.WriteString("\\section*{Skills}\n" + tex(strings.Join(doc.Skills, ", ")) + "\n")
			}
		case "experience":
			if len(doc.Experience) > 0 {
				b.WriteString("\\section*{Experience}\n")
			}
			for _, exp := range doc.Experience {
				b.WriteString("\\textbf{" + tex(exp.Title) + "}, " + tex(exp.Company) + " \\hfill " + tex(exp.Start) + "--" + tex(exp.End) + "\n\\begin{itemize}\n")
				for _, bullet := range exp.Bullets {
					b.WriteString("\\item " + tex(bullet.Text) + "\n")
				}
				b.WriteString("\\end{itemize}\n")
			}
		case "projects":
			if len(doc.Projects) > 0 {
				b.WriteString("\\section*{Projects}\n")
			}
			for _, project := range doc.Projects {
				b.WriteString("\\textbf{" + tex(project.Name) + "} -- " + tex(project.Description) + "\n\\begin{itemize}\n")
				for _, bullet := range project.Bullets {
					b.WriteString("\\item " + tex(bullet.Text) + "\n")
				}
				b.WriteString("\\end{itemize}\n")
			}
		case "education":
			if len(doc.Education) > 0 {
				b.WriteString("\\section*{Education}\n")
			}
			for _, edu := range doc.Education {
				b.WriteString("\\textbf{" + tex(edu.Institution) + "} -- " + tex(edu.Degree) + " \\hfill " + tex(edu.Start) + "--" + tex(edu.End) + "\n")
			}
		}
	}
	b.WriteString("\\end{document}\n")
	return b.String()
}

func tex(s string) string {
	replacer := strings.NewReplacer("\\", "\\textbackslash{}", "&", "\\&", "%", "\\%", "$", "\\$", "#", "\\#", "_", "\\_", "{", "\\{", "}", "\\}", "~", "\\textasciitilde{}", "^", "\\textasciicircum{}")
	return replacer.Replace(htmlstd.UnescapeString(s))
}
