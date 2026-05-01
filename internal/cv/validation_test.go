package cv

import (
	"strings"
	"testing"
)

func TestValidateRequiresName(t *testing.T) {
	doc := &CV{Summary: "summary", Projects: []Project{{Name: "p"}}}
	errs := Validate(doc)
	if len(errs) == 0 {
		t.Fatal("expected validation errors")
	}
}

func TestValidateVariantRejectsUnknownSection(t *testing.T) {
	variant := &Variant{Target: "backend engineer", SectionOrder: []string{"summary", "unknown"}}
	errs := ValidateVariant(variant)
	if len(errs) == 0 {
		t.Fatal("expected validation errors")
	}
}

func TestValidateVariantAcceptsExampleShape(t *testing.T) {
	variant := &Variant{
		Target:       "YC founder engineer",
		MaxPages:     1,
		Tone:         "direct",
		SectionOrder: []string{"summary", "experience", "projects", "skills", "education"},
	}
	if errs := ValidateVariant(variant); len(errs) > 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestWarningsDetectAgentRelevantIssues(t *testing.T) {
	duplicateBullet := "Built lead management workflows for cleaning operators."
	doc := &CV{
		Links: []Link{
			{Label: "GitHub", URL: ""},
		},
		Skills: []string{
			"Go", "Linux", "SQL", "Git", "Docker", "Kubernetes", "AWS", "GCP",
			"Terraform", "CI", "CD", "Bash", "Python", "TypeScript", "React", "YAML",
		},
		Experience: []Experience{
			{
				Bullets: []Bullet{
					{Text: "   "},
					{Text: duplicateBullet},
					{Text: "Improved conversion by 40%."},
					{Text: "[Sourced] Reduced support tickets by 20%."},
					{Text: "Built services with Go 1.22, API v2, and iOS 17 support."},
				},
			},
		},
		Projects: []Project{
			{Bullets: []Bullet{{Text: duplicateBullet}}},
			{}, {}, {}, {}, {}, {},
		},
	}

	warnings := Warnings(doc)

	assertWarning(t, warnings, "missing_link", "links[0]", "")
	assertWarning(t, warnings, "too_many_skills", "skills", "")
	assertWarning(t, warnings, "too_many_projects", "projects", "")
	assertWarning(t, warnings, "empty_bullet", "experience[0].bullets[0]", "")
	assertWarning(t, warnings, "duplicate_bullet", "projects[0].bullets[0]", "experience[0].bullets[1]")
	assertWarning(t, warnings, "suspicious_metric", "experience[0].bullets[2]", "")
	assertNoWarning(t, warnings, "suspicious_metric", "experience[0].bullets[3]")
	assertNoWarning(t, warnings, "suspicious_metric", "experience[0].bullets[4]")
}

func TestWarningsDetectLongBullets(t *testing.T) {
	doc := &CV{
		Projects: []Project{
			{Bullets: []Bullet{{Text: strings.Repeat("word ", 46)}}},
		},
	}

	warnings := Warnings(doc)

	assertWarning(t, warnings, "long_bullet", "projects[0].bullets[0]", "")
}

func assertWarning(t *testing.T, warnings []Warning, code, location, messageContains string) {
	t.Helper()
	for _, warning := range warnings {
		if warning.Code != code || warning.Location != location {
			continue
		}
		if messageContains != "" && !strings.Contains(warning.Message, messageContains) {
			t.Fatalf("expected warning %s at %s message to contain %q, got %q", code, location, messageContains, warning.Message)
		}
		return
	}
	t.Fatalf("expected warning %s at %s, got %#v", code, location, warnings)
}

func assertNoWarning(t *testing.T, warnings []Warning, code, location string) {
	t.Helper()
	for _, warning := range warnings {
		if warning.Code == code && warning.Location == location {
			t.Fatalf("unexpected warning %s at %s: %#v", code, location, warning)
		}
	}
}

func TestBulletUnmarshalJSONAcceptsStringAndObject(t *testing.T) {
	var scalar Bullet
	if err := scalar.UnmarshalJSON([]byte(`"Built workflow."`)); err != nil {
		t.Fatal(err)
	}
	if scalar.Text != "Built workflow." {
		t.Fatalf("unexpected scalar bullet: %#v", scalar)
	}

	var object Bullet
	if err := object.UnmarshalJSON([]byte(`{"text":"Interviewed operators.","source":"human","verified":true}`)); err != nil {
		t.Fatal(err)
	}
	if object.Text != "Interviewed operators." || object.Source != "human" || !object.Verified {
		t.Fatalf("unexpected object bullet: %#v", object)
	}
}
