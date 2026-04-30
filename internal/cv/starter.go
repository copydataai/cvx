package cv

import (
	"os"
	"path/filepath"
)

func WriteStarterFiles(root string) error {
	files := map[string]string{
		"cv.yaml":                           starterCV,
		"variants/yc-founder-engineer.yaml": ycFounderVariant,
		"variants/backend-engineer.yaml":    backendVariant,
		"variants/internship.yaml":          internshipVariant,
	}
	for path, body := range files {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if _, err := os.Stat(full); err == nil {
			continue
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

const starterCV = `name: Your Name
contact:
  email: you@example.com
  phone: ""
  location: City, Country
links:
  - label: GitHub
    url: https://github.com/yourname
summary: >
  Short factual professional summary. Do not include claims that are not supported by source data.
skills:
  - Go
  - Linux
experience:
  - company: Example Company
    title: Software Engineer
    location: Remote
    start: "2024"
    end: Present
    bullets:
      - Built and maintained backend services for internal users.
projects:
  - name: cvx
    description: Agent-native CV system.
    url: ""
    bullets:
      - Created structured YAML workflows for rendering CVs.
education:
  - institution: Example University
    degree: Example Degree
    start: "2020"
    end: "2024"
metadata:
  updated: "2026-04-30"
`

const ycFounderVariant = `target: YC founder engineer
max_pages: 1
tone: direct, technical, founder-oriented
section_order:
  - summary
  - experience
  - projects
  - skills
  - education
include_projects: []
exclude_projects: []
emphasis_keywords:
  - systems
  - product
  - AI
  - operations
  - infrastructure
`

const backendVariant = `target: backend engineer
max_pages: 1
tone: precise, systems-oriented, pragmatic
section_order:
  - summary
  - skills
  - experience
  - projects
  - education
include_projects: []
exclude_projects: []
emphasis_keywords:
  - APIs
  - databases
  - reliability
  - observability
  - infrastructure
`

const internshipVariant = `target: software engineering internship
max_pages: 1
tone: clear, evidence-driven, growth-oriented
section_order:
  - summary
  - education
  - projects
  - skills
  - experience
include_projects: []
exclude_projects: []
emphasis_keywords:
  - fundamentals
  - coursework
  - projects
  - GitHub
  - reliability
`
