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
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		usage()
		return nil
	}

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

func usage() {
	fmt.Printf(`Usage:
  cvx diff --from output/previous.json --to output/current.json [--output %s]

Compares two CV-shaped JSON snapshots and writes a markdown report.

Flags:
  --from    Required. Previous CV JSON snapshot.
  --to      Required. Current CV JSON snapshot.
  --output  Markdown report path. Default: %s
`, defaultOutput, defaultOutput)
}

func loadJSON(path string) (*cv.CV, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var snapshot struct {
		CV *cv.CV `json:"cv"`
	}
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if snapshot.CV != nil {
		return snapshot.CV, nil
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
	seen := map[string]int{}
	for i, item := range items {
		label := item.Company
		if label == "" {
			label = fmt.Sprintf("Experience %d", i+1)
		}
		if item.Title != "" {
			label = label + " - " + item.Title
		}
		key := stableKey(item.Title + " " + item.Company)
		if key == "" {
			key = stableKey(label)
		}
		key, label = occurrenceKeyAndLabel(key, label, seen)
		groups = append(groups, bulletGroup{
			key:     key,
			label:   label,
			bullets: bulletTexts(item.Bullets),
		})
	}
	return groups
}

func projectGroups(items []cv.Project) []bulletGroup {
	groups := make([]bulletGroup, 0, len(items))
	seen := map[string]int{}
	for i, item := range items {
		label := item.Name
		if label == "" {
			label = fmt.Sprintf("Project %d", i+1)
		}
		key := stableKey(item.Name)
		if key == "" {
			key = stableKey(label)
		}
		key, label = occurrenceKeyAndLabel(key, label, seen)
		groups = append(groups, bulletGroup{
			key:     key,
			label:   label,
			bullets: bulletTexts(item.Bullets),
		})
	}
	return groups
}

func bulletTexts(bullets []cv.Bullet) []string {
	texts := make([]string, 0, len(bullets))
	for _, bullet := range bullets {
		texts = append(texts, bullet.Text)
	}
	return texts
}

func occurrenceKeyAndLabel(key, label string, seen map[string]int) (string, string) {
	seen[key]++
	occurrence := seen[key]
	if occurrence == 1 {
		return key, label
	}
	return fmt.Sprintf("%s #%d", key, occurrence), fmt.Sprintf("%s #%d", label, occurrence)
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
	unmatchedFrom, unmatchedTo := unmatchedBullets(from, to)

	usedTo := make([]bool, len(unmatchedTo))
	for _, oldBullet := range unmatchedFrom {
		match := -1
		for j, newBullet := range unmatchedTo {
			if usedTo[j] {
				continue
			}
			if relatedBullets(oldBullet, newBullet) {
				match = j
				break
			}
		}
		if match == -1 {
			changes = append(changes, removedBulletChange(oldBullet))
			continue
		}
		usedTo[match] = true
		changes = append(changes,
			"- Changed wording\n"+
				fmt.Sprintf("  - Old bullet: %s\n", oldBullet)+
				fmt.Sprintf("  - New bullet: %s\n", unmatchedTo[match]),
		)
	}
	for j, newBullet := range unmatchedTo {
		if !usedTo[j] {
			changes = append(changes, addedBulletChange(newBullet))
		}
	}
	return changes
}

func unmatchedBullets(from, to []string) ([]string, []string) {
	toCounts := countBullets(to)
	unmatchedFrom := make([]string, 0, len(from))
	for _, bullet := range from {
		if toCounts[bullet] > 0 {
			toCounts[bullet]--
			continue
		}
		unmatchedFrom = append(unmatchedFrom, bullet)
	}

	fromCounts := countBullets(from)
	unmatchedTo := make([]string, 0, len(to))
	for _, bullet := range to {
		if fromCounts[bullet] > 0 {
			fromCounts[bullet]--
			continue
		}
		unmatchedTo = append(unmatchedTo, bullet)
	}
	return unmatchedFrom, unmatchedTo
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
	sharedLongToken := false
	for token := range fromTokens {
		if toTokens[token] {
			shared++
			if len(token) >= 8 {
				sharedLongToken = true
			}
		}
	}
	if shared >= 2 {
		return true
	}
	return shared == 1 && sharedLongToken && similarTokenCount(fromTokens, toTokens)
}

func similarTokenCount(from, to map[string]bool) bool {
	diff := len(from) - len(to)
	if diff < 0 {
		diff = -diff
	}
	return diff <= 2
}

func meaningfulTokens(text string) map[string]bool {
	tokens := map[string]bool{}
	for _, token := range wordPattern.FindAllString(strings.ToLower(text), -1) {
		if len(token) < 3 || stopwords[token] || genericActionVerbs[token] {
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

var genericActionVerbs = map[string]bool{
	"build":       true,
	"built":       true,
	"create":      true,
	"created":     true,
	"developed":   true,
	"implemented": true,
	"improved":    true,
	"led":         true,
	"managed":     true,
	"migrated":    true,
	"owned":       true,
	"supported":   true,
}

func countBullets(bullets []string) map[string]int {
	counts := make(map[string]int, len(bullets))
	for _, bullet := range bullets {
		counts[bullet]++
	}
	return counts
}
