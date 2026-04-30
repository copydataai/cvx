package diff

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/josesanchez/cvx/internal/cv"
)

const defaultOutput = "output/reports/last-diff.md"

// Command compares two CV-shaped JSON snapshots and writes a markdown report.
func Command(args []string) error {
	flags := flag.NewFlagSet("diff", flag.ContinueOnError)
	flags.SetOutput(new(bytes.Buffer))

	fromPath := flags.String("from", "", "previous CV JSON snapshot")
	toPath := flags.String("to", "", "current CV JSON snapshot")
	outputPath := flags.String("output", defaultOutput, "markdown diff report output path")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*fromPath) == "" {
		return fmt.Errorf("diff: --from is required")
	}
	if strings.TrimSpace(*toPath) == "" {
		return fmt.Errorf("diff: --to is required")
	}

	from, err := loadJSON(*fromPath)
	if err != nil {
		return err
	}
	to, err := loadJSON(*toPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		return fmt.Errorf("create output directory for %s: %w", *outputPath, err)
	}
	if err := os.WriteFile(*outputPath, []byte(Markdown(from, to)), 0o644); err != nil {
		return fmt.Errorf("write diff report %s: %w", *outputPath, err)
	}
	fmt.Println(*outputPath)
	return nil
}

func loadJSON(path string) (*cv.CV, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc cv.CV
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return &doc, nil
}

// Markdown returns a human-readable diff report for two CV snapshots.
func Markdown(from, to *cv.CV) string {
	var b strings.Builder
	b.WriteString("# CV Diff Report\n\n")
	writeChangedSections(&b, from, to)
	writeExperienceBullets(&b, from.Experience, to.Experience)
	writeProjectBullets(&b, from.Projects, to.Projects)
	return b.String()
}

func writeChangedSections(b *strings.Builder, from, to *cv.CV) {
	sections := changedSections(from, to)
	b.WriteString("## Changed Top-Level Sections\n\n")
	if len(sections) == 0 {
		b.WriteString("- None\n\n")
		return
	}
	for _, section := range sections {
		fmt.Fprintf(b, "- %s\n", section)
	}
	b.WriteString("\n")
}

func changedSections(from, to *cv.CV) []string {
	checks := []struct {
		name string
		old  any
		new  any
	}{
		{"Name", from.Name, to.Name},
		{"Contact", from.Contact, to.Contact},
		{"Links", from.Links, to.Links},
		{"Summary", from.Summary, to.Summary},
		{"Skills", from.Skills, to.Skills},
		{"Experience", from.Experience, to.Experience},
		{"Projects", from.Projects, to.Projects},
		{"Education", from.Education, to.Education},
		{"Metadata", from.Metadata, to.Metadata},
	}

	var changed []string
	for _, check := range checks {
		if !reflect.DeepEqual(check.old, check.new) {
			changed = append(changed, check.name)
		}
	}
	return changed
}

func writeExperienceBullets(b *strings.Builder, from, to []cv.Experience) {
	b.WriteString("## Experience Bullets\n\n")
	writeBulletGroups(b, experienceGroups(from), experienceGroups(to))
}

func writeProjectBullets(b *strings.Builder, from, to []cv.Project) {
	b.WriteString("## Projects Bullets\n\n")
	writeBulletGroups(b, projectGroups(from), projectGroups(to))
}

type bulletGroup struct {
	label   string
	bullets []string
}

func experienceGroups(items []cv.Experience) []bulletGroup {
	groups := make([]bulletGroup, 0, len(items))
	for i, item := range items {
		label := item.Company
		if label == "" {
			label = fmt.Sprintf("Experience %d", i+1)
		}
		if item.Title != "" {
			label = label + " - " + item.Title
		}
		groups = append(groups, bulletGroup{label: label, bullets: item.Bullets})
	}
	return groups
}

func projectGroups(items []cv.Project) []bulletGroup {
	groups := make([]bulletGroup, 0, len(items))
	for i, item := range items {
		label := item.Name
		if label == "" {
			label = fmt.Sprintf("Project %d", i+1)
		}
		groups = append(groups, bulletGroup{label: label, bullets: item.Bullets})
	}
	return groups
}

func writeBulletGroups(b *strings.Builder, from, to []bulletGroup) {
	wrote := false
	max := max(len(from), len(to))
	for i := 0; i < max; i++ {
		var oldGroup, newGroup bulletGroup
		if i < len(from) {
			oldGroup = from[i]
		}
		if i < len(to) {
			newGroup = to[i]
		}
		label := newGroup.label
		if label == "" {
			label = oldGroup.label
		}
		changes := bulletChanges(oldGroup.bullets, newGroup.bullets)
		if len(changes) == 0 {
			continue
		}
		wrote = true
		fmt.Fprintf(b, "### %s\n\n", label)
		for _, change := range changes {
			b.WriteString(change)
		}
		b.WriteString("\n")
	}
	if !wrote {
		b.WriteString("- None\n\n")
	}
}

func bulletChanges(from, to []string) []string {
	var changes []string
	if len(from) == len(to) {
		for i := range from {
			if from[i] == to[i] {
				continue
			}
			changes = append(changes,
				"- Changed wording\n"+
					fmt.Sprintf("  - Old bullet: %s\n", from[i])+
					fmt.Sprintf("  - New bullet: %s\n", to[i]),
			)
		}
		return changes
	}

	oldCounts := countBullets(from)
	newCounts := countBullets(to)
	for _, bullet := range from {
		if oldCounts[bullet] > newCounts[bullet] {
			changes = append(changes, fmt.Sprintf("- Removed bullet: %s\n", bullet))
			oldCounts[bullet]--
		}
	}
	oldCounts = countBullets(from)
	for _, bullet := range to {
		if newCounts[bullet] > oldCounts[bullet] {
			changes = append(changes, fmt.Sprintf("- Added bullet: %s\n", bullet))
			newCounts[bullet]--
		}
	}
	return changes
}

func countBullets(bullets []string) map[string]int {
	counts := make(map[string]int, len(bullets))
	for _, bullet := range bullets {
		counts[bullet]++
	}
	return counts
}
