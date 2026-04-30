package diff

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/josesanchez/cvx/internal/cv"
)

const defaultOutput = "output/reports/last-diff.md"

var wordPattern = regexp.MustCompile(`[a-z0-9]+`)

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
	key     string
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
		groups = append(groups, bulletGroup{
			key:     stableKey(item.Title + " " + item.Company),
			label:   label,
			bullets: item.Bullets,
		})
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
		groups = append(groups, bulletGroup{
			key:     stableKey(item.Name),
			label:   label,
			bullets: item.Bullets,
		})
	}
	return groups
}

func writeBulletGroups(b *strings.Builder, from, to []bulletGroup) {
	wrote := false
	fromByKey := groupByKey(from)
	toByKey := groupByKey(to)
	seen := map[string]bool{}

	for _, oldGroup := range from {
		newGroup, ok := toByKey[oldGroup.key]
		changes := removedBulletChanges(oldGroup.bullets)
		if ok {
			changes = bulletChanges(oldGroup.bullets, newGroup.bullets)
		}
		if len(changes) == 0 {
			seen[oldGroup.key] = true
			continue
		}
		wrote = true
		fmt.Fprintf(b, "### %s\n\n", oldGroup.label)
		for _, change := range changes {
			b.WriteString(change)
		}
		b.WriteString("\n")
		seen[oldGroup.key] = true
	}

	for _, newGroup := range to {
		if seen[newGroup.key] {
			continue
		}
		if _, ok := fromByKey[newGroup.key]; ok {
			continue
		}
		changes := addedBulletChanges(newGroup.bullets)
		if len(changes) == 0 {
			continue
		}
		wrote = true
		fmt.Fprintf(b, "### %s\n\n", newGroup.label)
		for _, change := range changes {
			b.WriteString(change)
		}
		b.WriteString("\n")
	}
	if !wrote {
		b.WriteString("- None\n\n")
	}
}

func groupByKey(groups []bulletGroup) map[string]bulletGroup {
	byKey := make(map[string]bulletGroup, len(groups))
	for _, group := range groups {
		byKey[group.key] = group
	}
	return byKey
}

func bulletChanges(from, to []string) []string {
	var changes []string
	if len(from) == len(to) {
		for i := range from {
			if from[i] == to[i] {
				continue
			}
			if relatedBullets(from[i], to[i]) {
				changes = append(changes,
					"- Changed wording\n"+
						fmt.Sprintf("  - Old bullet: %s\n", from[i])+
						fmt.Sprintf("  - New bullet: %s\n", to[i]),
				)
				continue
			}
			changes = append(changes, fmt.Sprintf("- Removed bullet: %s\n", from[i]))
			changes = append(changes, fmt.Sprintf("- Added bullet: %s\n", to[i]))
		}
		return changes
	}

	oldCounts := countBullets(from)
	newCounts := countBullets(to)
	for _, bullet := range from {
		if oldCounts[bullet] > newCounts[bullet] {
			changes = append(changes, removedBulletChange(bullet))
			oldCounts[bullet]--
		}
	}
	oldCounts = countBullets(from)
	for _, bullet := range to {
		if newCounts[bullet] > oldCounts[bullet] {
			changes = append(changes, addedBulletChange(bullet))
			newCounts[bullet]--
		}
	}
	return changes
}

func removedBulletChanges(bullets []string) []string {
	changes := make([]string, 0, len(bullets))
	for _, bullet := range bullets {
		changes = append(changes, removedBulletChange(bullet))
	}
	return changes
}

func addedBulletChanges(bullets []string) []string {
	changes := make([]string, 0, len(bullets))
	for _, bullet := range bullets {
		changes = append(changes, addedBulletChange(bullet))
	}
	return changes
}

func removedBulletChange(bullet string) string {
	return fmt.Sprintf("- Removed bullet: %s\n", bullet)
}

func addedBulletChange(bullet string) string {
	return fmt.Sprintf("- Added bullet: %s\n", bullet)
}

func relatedBullets(from, to string) bool {
	fromTokens := meaningfulTokens(from)
	toTokens := meaningfulTokens(to)
	if len(fromTokens) == 0 || len(toTokens) == 0 {
		return false
	}

	shared := 0
	for token := range fromTokens {
		if toTokens[token] {
			shared++
		}
	}
	return shared >= 1
}

func meaningfulTokens(text string) map[string]bool {
	tokens := map[string]bool{}
	for _, token := range wordPattern.FindAllString(strings.ToLower(text), -1) {
		if len(token) < 3 || stopwords[token] {
			continue
		}
		tokens[token] = true
	}
	return tokens
}

func stableKey(text string) string {
	key := strings.Join(wordPattern.FindAllString(strings.ToLower(text), -1), " ")
	if key == "" {
		return strings.ToLower(strings.TrimSpace(text))
	}
	return key
}

var stopwords = map[string]bool{
	"and":  true,
	"for":  true,
	"from": true,
	"into": true,
	"the":  true,
	"this": true,
	"that": true,
	"with": true,
}

func countBullets(bullets []string) map[string]int {
	counts := make(map[string]int, len(bullets))
	for _, bullet := range bullets {
		counts[bullet]++
	}
	return counts
}
