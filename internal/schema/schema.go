package schema

import (
	"os"
	"path/filepath"
)

const cvSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/josesanchez/cvx/schema/cv.schema.json",
  "title": "CV",
  "type": "object",
  "additionalProperties": false,
  "required": ["name", "summary"],
  "properties": {
    "name": {"type": "string"},
    "contact": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "email": {"type": "string"},
        "phone": {"type": "string"},
        "location": {"type": "string"}
      }
    },
    "links": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "properties": {
          "label": {"type": "string"},
          "url": {"type": "string"}
        }
      }
    },
    "summary": {"type": "string"},
    "skills": {
      "type": "array",
      "items": {"type": "string"}
    },
    "experience": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "properties": {
          "company": {"type": "string"},
          "title": {"type": "string"},
          "location": {"type": "string"},
          "start": {"type": "string"},
          "end": {"type": "string"},
          "bullets": {
            "type": "array",
            "items": {"type": "string"}
          }
        }
      }
    },
    "projects": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "properties": {
          "name": {"type": "string"},
          "description": {"type": "string"},
          "url": {"type": "string"},
          "bullets": {
            "type": "array",
            "items": {"type": "string"}
          }
        }
      }
    },
    "education": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "properties": {
          "institution": {"type": "string"},
          "degree": {"type": "string"},
          "start": {"type": "string"},
          "end": {"type": "string"}
        }
      }
    },
    "metadata": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "updated": {"type": "string"}
      }
    }
  }
}
`

const variantSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/josesanchez/cvx/schema/variant.schema.json",
  "title": "CV Variant",
  "type": "object",
  "additionalProperties": false,
  "required": ["target", "section_order"],
  "properties": {
    "target": {"type": "string"},
    "max_pages": {
      "type": "integer",
      "minimum": 0
    },
    "tone": {"type": "string"},
    "section_order": {
      "type": "array",
      "items": {
        "type": "string",
        "enum": ["summary", "experience", "projects", "skills", "education"]
      }
    },
    "include_projects": {
      "type": "array",
      "items": {"type": "string"}
    },
    "exclude_projects": {
      "type": "array",
      "items": {"type": "string"}
    },
    "emphasis_keywords": {
      "type": "array",
      "items": {"type": "string"}
    }
  }
}
`

// WriteSchemas writes the static CV and variant JSON schemas into dir.
func WriteSchemas(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	files := map[string]string{
		"cv.schema.json":      cvSchema,
		"variant.schema.json": variantSchema,
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			return err
		}
	}
	return nil
}
