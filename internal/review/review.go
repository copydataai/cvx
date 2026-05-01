package review

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/josesanchez/cvx/internal/cv"
)

var defaultSections = []string{"summary", "experience", "projects", "skills", "education"}

func Command(args []string) error {
	if len(args) == 0 {
		usage()
		return nil
	}
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		usage()
		return nil
	}

	subcommand := args[0]
	switch subcommand {
	case "facts", "bullets", "ats", "target":
		return run(subcommand, args[1:])
	default:
		return fmt.Errorf("review: unknown subcommand %q", subcommand)
	}
}

func run(name string, args []string) error {
	fs := flag.NewFlagSet("review "+name, flag.ContinueOnError)
	input := fs.String("input", "cv.yaml", "CV YAML path")
	variantPath := fs.String("variant", "", "variant YAML path")
	output := fs.String("output", filepath.Join("output", "reports", "review-"+name+".md"), "review markdown output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("review %s: unexpected argument %q", name, fs.Arg(0))
	}
	if name == "target" && strings.TrimSpace(*variantPath) == "" {
		return fmt.Errorf("review target: --variant is required")
	}

	report, err := Build(name, *input, *variantPath)
	if err != nil {
		return err
	}
	if err := writeMarkdown(*output, report); err != nil {
		return err
	}
	fmt.Println(*output)
	return nil
}

type Report struct {
	Name     string
	Title    string
	Sections []Section
}

type Section struct {
	Heading string
	Items   []string
}

func Build(name, input, variantPath string) (*Report, error) {
	doc, err := cv.Load(input)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", input, err)
	}
	warnings := cv.Warnings(doc)

	switch name {
	case "facts":
		return factsReport(doc, warnings), nil
	case "bullets":
		return bulletsReport(doc, warnings), nil
	case "ats":
		return atsReport(doc, warnings), nil
	case "target":
		if strings.TrimSpace(variantPath) == "" {
			return nil, fmt.Errorf("review target: --variant is required")
		}
		variant, err := cv.LoadVariant(variantPath)
		if err != nil {
			return nil, fmt.Errorf("load variant %s: %w", variantPath, err)
		}
		return targetReport(doc, variant), nil
	default:
		return nil, fmt.Errorf("review: unknown subcommand %q", name)
	}
}

func factsReport(doc *cv.CV, warnings []cv.Warning) *Report {
	return &Report{
		Name:  doc.Name,
		Title: "Facts",
		Sections: []Section{{
			Heading: "Warnings",
			Items:   warningItems(warnings, "suspicious_metric", "missing_link"),
		}},
	}
}

func bulletsReport(doc *cv.CV, warnings []cv.Warning) *Report {
	return &Report{
		Name:  doc.Name,
		Title: "Bullets",
		Sections: []Section{{
			Heading: "Warnings",
			Items:   warningItems(warnings, "empty_bullet", "long_bullet", "duplicate_bullet"),
		}},
	}
}

func atsReport(doc *cv.CV, warnings []cv.Warning) *Report {
	return &Report{
		Name:  doc.Name,
		Title: "ATS",
		Sections: []Section{{
			Heading: "Warnings",
			Items:   warningItems(warnings, "too_many_skills", "too_many_projects", "missing_link"),
		}},
	}
}

func targetReport(doc *cv.CV, variant *cv.Variant) *Report {
	sectionOrder := variant.SectionOrder
	if len(sectionOrder) == 0 {
		sectionOrder = defaultSections
	}

	return &Report{
		Name:  doc.Name,
		Title: "Target",
		Sections: []Section{
			{Heading: "Target", Items: valueOrNone(variant.Target)},
			{Heading: "Section Order", Items: listOrNone(sectionOrder)},
			{Heading: "Emphasis Keywords", Items: listOrNone(variant.EmphasisKeywords)},
			{Heading: "Included Projects", Items: listOrNone(variant.IncludeProjects)},
			{Heading: "Excluded Projects", Items: listOrNone(variant.ExcludeProjects)},
			{Heading: "Missing Included Projects", Items: missingIncludedProjects(doc, variant)},
		},
	}
}

func warningItems(warnings []cv.Warning, codes ...string) []string {
	allowed := map[string]bool{}
	for _, code := range codes {
		allowed[code] = true
	}
	items := []string{}
	for _, warning := range warnings {
		if !allowed[warning.Code] {
			continue
		}
		items = append(items, fmt.Sprintf("%s at %s: %s", warning.Code, warning.Location, warning.Message))
	}
	if len(items) == 0 {
		return []string{"No warnings found."}
	}
	return items
}

func missingIncludedProjects(doc *cv.CV, variant *cv.Variant) []string {
	projects := map[string]bool{}
	for _, project := range doc.Projects {
		projects[project.Name] = true
	}
	missing := []string{}
	for _, name := range variant.IncludeProjects {
		if !projects[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) == 0 {
		return []string{"No missing included projects."}
	}
	return missing
}

func listOrNone(items []string) []string {
	if len(items) == 0 {
		return []string{"None."}
	}
	return append([]string(nil), items...)
}

func valueOrNone(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{"None."}
	}
	return []string{value}
}

func writeMarkdown(path string, report *Report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(markdown(report)), 0o644)
}

func markdown(report *Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# CV %s Review\n\n", report.Name)
	fmt.Fprintf(&b, "## %s\n\n", report.Title)
	for _, section := range report.Sections {
		fmt.Fprintf(&b, "### %s\n", section.Heading)
		for _, item := range section.Items {
			fmt.Fprintf(&b, "- %s\n", item)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func usage() {
	fmt.Print(`Usage:
  cvx review facts [--input cv.yaml] [--output output/reports/review-facts.md]
  cvx review bullets [--input cv.yaml] [--output output/reports/review-bullets.md]
  cvx review ats [--input cv.yaml] [--output output/reports/review-ats.md]
  cvx review target --variant variants/name.yaml [--input cv.yaml] [--output output/reports/review-target.md]

Writes deterministic Markdown review reports without AI.
`)
}
