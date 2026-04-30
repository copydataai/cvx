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

func Warnings(doc *CV) []Warning {
	var warnings []Warning

	for i, link := range doc.Links {
		if strings.TrimSpace(link.URL) == "" {
			warnings = append(warnings, Warning{
				Code:     "missing_link",
				Message:  "link URL is empty",
				Location: fmt.Sprintf("links[%d]", i),
			})
		}
	}

	if len(doc.Skills) > 15 {
		warnings = append(warnings, Warning{
			Code:     "too_many_skills",
			Message:  "more than 15 skills may reduce focus",
			Location: "skills",
		})
	}

	if len(doc.Projects) > 6 {
		warnings = append(warnings, Warning{
			Code:     "too_many_projects",
			Message:  "more than 6 projects may reduce focus",
			Location: "projects",
		})
	}

	seenBullets := map[string]string{}
	for i, exp := range doc.Experience {
		for j, bullet := range exp.Bullets {
			location := fmt.Sprintf("experience[%d].bullets[%d]", i, j)
			warnings = append(warnings, bulletWarnings(bullet, location, seenBullets)...)
		}
	}
	for i, project := range doc.Projects {
		for j, bullet := range project.Bullets {
			location := fmt.Sprintf("projects[%d].bullets[%d]", i, j)
			warnings = append(warnings, bulletWarnings(bullet, location, seenBullets)...)
		}
	}

	return warnings
}

func bulletWarnings(bullet, location string, seenBullets map[string]string) []Warning {
	var warnings []Warning
	trimmed := strings.TrimSpace(bullet)

	if trimmed == "" {
		warnings = append(warnings, Warning{
			Code:     "empty_bullet",
			Message:  "bullet is empty",
			Location: location,
		})
		return warnings
	}

	if len(strings.Fields(trimmed)) > 45 {
		warnings = append(warnings, Warning{
			Code:     "long_bullet",
			Message:  "bullet is over 45 words",
			Location: location,
		})
	}

	if firstLocation, ok := seenBullets[trimmed]; ok {
		warnings = append(warnings, Warning{
			Code:     "duplicate_bullet",
			Message:  fmt.Sprintf("duplicate bullet also appears at %s", firstLocation),
			Location: location,
		})
	} else {
		seenBullets[trimmed] = location
	}

	if containsSuspiciousMetric(trimmed) {
		warnings = append(warnings, Warning{
			Code:     "suspicious_metric",
			Message:  "bullet contains a metric without [Sourced] or [Verified]",
			Location: location,
		})
	}

	return warnings
}

func containsSuspiciousMetric(bullet string) bool {
	if strings.Contains(bullet, "[Sourced]") || strings.Contains(bullet, "[Verified]") {
		return false
	}
	if strings.Contains(bullet, "%") {
		return true
	}
	for _, r := range bullet {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
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
