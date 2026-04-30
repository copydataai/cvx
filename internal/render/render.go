package render

import (
	"flag"
	"fmt"
	"html"
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
	output := fs.String("output", "output/cv.tex", "output TeX path")
	reportPath := fs.String("report", "output/reports/last-render.json", "render report JSON path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	sections := []string{"summary", "experience", "projects", "skills", "education"}
	doc, err := cv.Load(*input)
	if err != nil {
		return err
	}
	warnings := reportWarnings(cv.Warnings(doc))
	renderReport := newRenderReport(*input, *output, *variantPath, sections, warnings)
	if errs := cv.Validate(doc); len(errs) > 0 {
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
			renderReport.Validation = report.ValidationResult{Success: false, Errors: errs}
			if err := report.WriteRenderReport(*reportPath, renderReport); err != nil {
				return fmt.Errorf("write render report: %w", err)
			}
			return fmt.Errorf("variant validation failed: %s", strings.Join(errs, "; "))
		}
		sections = variant.SectionOrder
		renderReport.SectionOrder = append([]string(nil), sections...)
		doc.Projects = applyProjectFilter(doc.Projects, variant)
	}

	body := renderTeX(doc, sections)
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(*output, []byte(body), 0o644); err != nil {
		return err
	}
	renderReport.Validation = report.ValidationResult{Success: true}
	if err := report.WriteRenderReport(*reportPath, renderReport); err != nil {
		return err
	}
	fmt.Println("rendered", *output)
	return nil
}

func newRenderReport(input, output, variantPath string, sections []string, warnings []report.ReportWarning) report.RenderReport {
	return report.RenderReport{
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Input:        input,
		OutputTeX:    output,
		OutputPDF:    outputPDFPath(output),
		Variant:      variantPath,
		Engine:       "tex-only",
		Warnings:     warnings,
		SectionOrder: append([]string(nil), sections...),
	}
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

func applyProjectFilter(projects []cv.Project, variant *cv.Variant) []cv.Project {
	include := toSet(variant.IncludeProjects)
	exclude := toSet(variant.ExcludeProjects)
	var filtered []cv.Project
	for _, project := range projects {
		_, explicitlyIncluded := include[project.Name]
		if len(include) > 0 && !explicitlyIncluded {
			continue
		}
		if exclude[project.Name] {
			continue
		}
		filtered = append(filtered, project)
	}
	return filtered
}

func toSet(items []string) map[string]bool {
	set := map[string]bool{}
	for _, item := range items {
		set[item] = true
	}
	return set
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
					b.WriteString("\\item " + tex(bullet) + "\n")
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
					b.WriteString("\\item " + tex(bullet) + "\n")
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
	return replacer.Replace(html.UnescapeString(s))
}
