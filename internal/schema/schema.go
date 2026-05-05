package schema

import (
	"fmt"
	"os"
	"path/filepath"
)

const bulletSchema = `{
  "oneOf": [
    {"type": "string"},
    {
      "type": "object",
      "additionalProperties": false,
      "required": ["text"],
      "properties": {
        "text": {"type": "string"},
        "source": {"type": "string"},
        "sources": {"type": "array", "items": {"type": "string"}},
        "verified": {"type": "boolean"}
      }
    }
  ]
}`

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
    "sources": {"type": "array", "items": {"$ref": "#/$defs/source"}},
    "summary": {"type": "string"},
    "skills": {"type": "array", "items": {"type": "string"}},
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
          "bullets": {"type": "array", "items": {"$ref": "#/$defs/bullet"}}
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
          "bullets": {"type": "array", "items": {"$ref": "#/$defs/bullet"}}
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
      "properties": {"updated": {"type": "string"}}
    }
  },
  "$defs": {
    "bullet": ` + bulletSchema + `,
    "source": {
      "type": "object",
      "additionalProperties": false,
      "required": ["id", "type"],
      "properties": {
        "id": {"type": "string"},
        "type": {"type": "string"},
        "label": {"type": "string"},
        "url": {"type": "string"},
        "notes": {"type": "string"}
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
    "max_pages": {"type": "integer", "minimum": 0},
    "tone": {"type": "string"},
    "section_order": {
      "type": "array",
      "items": {"type": "string", "enum": ["summary", "experience", "projects", "skills", "education"]}
    },
    "include_projects": {"type": "array", "items": {"type": "string"}},
    "exclude_projects": {"type": "array", "items": {"type": "string"}},
    "emphasis_keywords": {"type": "array", "items": {"type": "string"}}
  }
}
`

const snapshotSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/josesanchez/cvx/schema/snapshot.schema.json",
  "title": "CV Snapshot",
  "type": "object",
  "additionalProperties": false,
  "required": ["input", "section_order", "cv"],
  "properties": {
    "input": {"type": "string"},
    "variant": {"type": "string"},
    "target": {"type": "string"},
    "section_order": {"type": "array", "items": {"type": "string"}},
    "cv": {"$ref": "cv.schema.json"}
  }
}
`

// WriteSchemas writes the static JSON schemas into dir.
func WriteSchemas(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	files := map[string]string{
		"cv.schema.json":       cvSchema,
		"variant.schema.json":  variantSchema,
		"snapshot.schema.json": snapshotSchema,
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func CheckSchemas(dir string) error {
	tmp, err := os.MkdirTemp("", "cvx-schema-check-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if err := WriteSchemas(tmp); err != nil {
		return err
	}
	for _, name := range []string{"cv.schema.json", "variant.schema.json", "snapshot.schema.json"} {
		current, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		expected, err := os.ReadFile(filepath.Join(tmp, name))
		if err != nil {
			return err
		}
		if string(current) != string(expected) {
			return fmt.Errorf("%s differs from generated schema; run `cvx schema`", filepath.Join(dir, name))
		}
	}
	return nil
}
