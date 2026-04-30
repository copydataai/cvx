package cv

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var validSections = map[string]bool{
	"summary": true, "experience": true, "projects": true, "skills": true, "education": true,
}

func Load(path string) (*CV, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc CV
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func LoadVariant(path string) (*Variant, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var variant Variant
	if err := yaml.Unmarshal(data, &variant); err != nil {
		return nil, err
	}
	return &variant, nil
}

func Validate(doc *CV) []string {
	var errs []string
	if strings.TrimSpace(doc.Name) == "" {
		errs = append(errs, "name is required")
	}
	if strings.TrimSpace(doc.Summary) == "" {
		errs = append(errs, "summary is required")
	}
	if len(doc.Experience) == 0 && len(doc.Projects) == 0 {
		errs = append(errs, "at least one experience or project is required")
	}
	return errs
}

func ValidateVariant(v *Variant) []string {
	var errs []string
	if strings.TrimSpace(v.Target) == "" {
		errs = append(errs, "variant target is required")
	}
	if v.MaxPages < 0 {
		errs = append(errs, "variant max_pages cannot be negative")
	}
	seen := map[string]bool{}
	for _, section := range v.SectionOrder {
		if !validSections[section] {
			errs = append(errs, fmt.Sprintf("unknown section %q in variant section_order", section))
		}
		if seen[section] {
			errs = append(errs, fmt.Sprintf("duplicate section %q in variant section_order", section))
		}
		seen[section] = true
	}
	return errs
}
