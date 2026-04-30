package cv

import "testing"

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
