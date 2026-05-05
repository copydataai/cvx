package policy

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/josesanchez/cvx/internal/cv"
)

type Report struct {
	Status   string  `json:"status"`
	Approved bool    `json:"approved"`
	Risks    []Risk  `json:"risks"`
	Summary  Summary `json:"summary"`
}

type Summary struct {
	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
}

type Risk struct {
	Level    string `json:"level"`
	Code     string `json:"code"`
	Location string `json:"location"`
	Message  string `json:"message"`
}

func Command(args []string) error {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		usage()
		return nil
	}
	fs := flag.NewFlagSet("policy", flag.ContinueOnError)
	before := fs.String("before", "", "before snapshot JSON")
	after := fs.String("after", "", "after snapshot JSON")
	output := fs.String("output", "output/reports/policy-report.json", "policy report JSON path")
	approveHigh := fs.Bool("approve-high", false, "allow high-risk changes")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*before) == "" || strings.TrimSpace(*after) == "" {
		return fmt.Errorf("policy: --before and --after are required")
	}
	report, err := Check(*before, *after, *approveHigh)
	if err != nil {
		return err
	}
	if err := writeReport(*output, report); err != nil {
		return err
	}
	fmt.Println(*output)
	if report.Summary.High > 0 && !report.Approved {
		return fmt.Errorf("policy blocked %d high-risk change(s); rerun with --approve-high after human approval", report.Summary.High)
	}
	return nil
}

func Check(beforePath, afterPath string, approved bool) (*Report, error) {
	before, err := loadSnapshot(beforePath)
	if err != nil {
		return nil, err
	}
	after, err := loadSnapshot(afterPath)
	if err != nil {
		return nil, err
	}
	risks := risks(before, after)
	report := &Report{Status: "pass", Approved: approved, Risks: risks}
	for _, risk := range risks {
		switch risk.Level {
		case "High":
			report.Summary.High++
		case "Medium":
			report.Summary.Medium++
		default:
			report.Summary.Low++
		}
	}
	if report.Summary.High > 0 && !approved {
		report.Status = "blocked"
	}
	return report, nil
}

func risks(before, after *cv.CV) []Risk {
	var risks []Risk
	if before.Name != after.Name {
		risks = append(risks, high("change_name", "name", "Name changed."))
	}
	if !reflect.DeepEqual(before.Contact, after.Contact) {
		risks = append(risks, high("change_contact", "contact", "Contact details changed."))
	}
	if !reflect.DeepEqual(before.Education, after.Education) {
		risks = append(risks, high("change_education", "education", "Education changed."))
	}
	for i, exp := range after.Experience {
		if i >= len(before.Experience) {
			risks = append(risks, high("add_experience", fmt.Sprintf("experience[%d]", i), "Experience entry added."))
			continue
		}
		old := before.Experience[i]
		loc := fmt.Sprintf("experience[%d]", i)
		if old.Company != exp.Company || old.Title != exp.Title || old.Start != exp.Start || old.End != exp.End {
			risks = append(risks, high("change_experience_identity", loc, "Company, title, or dates changed."))
		}
	}
	risks = append(risks, bulletRisks(after)...)
	return risks
}

func bulletRisks(doc *cv.CV) []Risk {
	var risks []Risk
	for i, exp := range doc.Experience {
		for j, bullet := range exp.Bullets {
			risks = appendBulletRisk(risks, fmt.Sprintf("experience[%d].bullets[%d]", i, j), bullet)
		}
	}
	for i, project := range doc.Projects {
		for j, bullet := range project.Bullets {
			risks = appendBulletRisk(risks, fmt.Sprintf("projects[%d].bullets[%d]", i, j), bullet)
		}
	}
	return risks
}

func appendBulletRisk(risks []Risk, loc string, bullet cv.Bullet) []Risk {
	if strings.TrimSpace(bullet.Text) == "" {
		return risks
	}
	if !bullet.Verified {
		risks = append(risks, Risk{Level: "Medium", Code: "unverified_bullet", Location: loc, Message: "Bullet is not marked verified."})
	}
	if strings.TrimSpace(bullet.Source) == "" && len(bullet.Sources) == 0 {
		risks = append(risks, Risk{Level: "Medium", Code: "missing_bullet_source", Location: loc, Message: "Bullet has no source reference."})
	}
	return risks
}

func high(code, location, message string) Risk {
	return Risk{Level: "High", Code: code, Location: location, Message: message}
}

func loadSnapshot(path string) (*cv.CV, error) {
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

func writeReport(path string, report *Report) error {
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

func usage() {
	fmt.Print(`Usage:
  cvx policy --before output/previous.json --after output/current.json [--approve-high]

Checks an agent edit transaction for high-risk factual changes.
`)
}
