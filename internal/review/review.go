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
		Sections: []Section{
			{Heading: "Warnings", Items: warningItems(warnings, "suspicious_metric", "missing_link")},
			{Heading: "Provenance", Items: provenanceItems(doc)},
			{Heading: "Source Registry", Items: sourceRegistryItems(doc)},
			{Heading: "Actions", Items: []string{"Verify any metric warnings with source data before using them.", "Add bullet source/verified fields for high-stakes claims."}},
		},
	}
}

func bulletsReport(doc *cv.CV, warnings []cv.Warning) *Report {
	return &Report{
		Name:  doc.Name,
		Title: "Bullets",
		Sections: []Section{
			{Heading: "Warnings", Items: warningItems(warnings, "empty_bullet", "long_bullet", "duplicate_bullet")},
			{Heading: "Actions", Items: []string{"Remove empty bullets.", "Split bullets longer than 45 words.", "Merge or delete duplicate bullets."}},
		},
	}
}

func atsReport(doc *cv.CV, warnings []cv.Warning) *Report {
	return &Report{
		Name:  doc.Name,
		Title: "ATS",
		Sections: []Section{
			{Heading: "Warnings", Items: warningItems(warnings, "too_many_skills", "too_many_projects", "missing_link")},
			{Heading: "Actions", Items: []string{"Keep section names conventional.", "Prefer plain bullets over visual-only formatting.", "Keep skills focused on target-relevant technologies."}},
		},
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
			{Heading: "Actions", Items: []string{"Check that emphasis keywords are supported by cv.yaml facts.", "Prefer excluding weak projects over adding unsupported claims."}},
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
		items = append(items, fmt.Sprintf("[%s] %s at %s: %s", severity(warning.Code), warning.Code, warning.Location, warning.Message))
	}
	if len(items) == 0 {
		return []string{"[OK] No warnings found."}
	}
	return items
}

func severity(code string) string {
	switch code {
	case "suspicious_metric", "missing_link", "empty_bullet":
		return "High"
	case "duplicate_bullet", "long_bullet":
		return "Medium"
	default:
		return "Low"
	}
}

func provenanceItems(doc *cv.CV) []string {
	items := []string{}
	for i, exp := range doc.Experience {
		for j, bullet := range exp.Bullets {
			location := fmt.Sprintf("experience[%d].bullets[%d]", i, j)
			items = appendProvenanceItem(items, location, bullet)
		}
	}
	for i, project := range doc.Projects {
		for j, bullet := range project.Bullets {
			location := fmt.Sprintf("projects[%d].bullets[%d]", i, j)
			items = appendProvenanceItem(items, location, bullet)
		}
	}
	if len(items) == 0 {
		return []string{"[OK] All bullets include source and verified provenance, or no bullets exist."}
	}
	return items
}

func appendProvenanceItem(items []string, location string, bullet cv.Bullet) []string {
	if strings.TrimSpace(bullet.Source) == "" && len(bullet.Sources) == 0 {
		items = append(items, fmt.Sprintf("[Medium] %s has no source field.", location))
	}
	if !bullet.Verified {
		items = append(items, fmt.Sprintf("[Medium] %s is not marked verified.", location))
	}
	return items
}

func sourceRegistryItems(doc *cv.CV) []string {
	registry := map[string]bool{}
	for _, source := range doc.Sources {
		if strings.TrimSpace(source.ID) != "" {
			registry[source.ID] = true
		}
	}
	items := []string{}
	for i, exp := range doc.Experience {
		for j, bullet := range exp.Bullets {
			items = appendSourceReferenceItems(items, fmt.Sprintf("experience[%d].bullets[%d]", i, j), bullet, registry)
		}
	}
	for i, project := range doc.Projects {
		for j, bullet := range project.Bullets {
			items = appendSourceReferenceItems(items, fmt.Sprintf("projects[%d].bullets[%d]", i, j), bullet, registry)
		}
	}
	if len(items) == 0 {
		return []string{"[OK] Bullet source references resolve."}
	}
	return items
}

func appendSourceReferenceItems(items []string, location string, bullet cv.Bullet, registry map[string]bool) []string {
	for _, id := range bullet.Sources {
		if !registry[id] {
			items = append(items, fmt.Sprintf("[High] %s references missing source %q.", location, id))
		}
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
			missing = append(missing, fmt.Sprintf("[High] %s", name))
		}
	}
	if len(missing) == 0 {
		return []string{"[OK] No missing included projects."}
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
