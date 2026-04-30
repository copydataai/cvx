package schema

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteSchemas(t *testing.T) {
	dir := t.TempDir()
	if err := WriteSchemas(dir); err != nil {
		t.Fatalf("WriteSchemas() error = %v", err)
	}

	for _, name := range []string{"cv.schema.json", "variant.schema.json"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}

		var decoded map[string]any
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}
		if decoded["$schema"] == "" {
			t.Fatalf("%s missing $schema", name)
		}
		if decoded["type"] != "object" {
			t.Fatalf("%s type = %v, want object", name, decoded["type"])
		}
	}
}

func TestCheckedInSchemasMatchGeneratedOutput(t *testing.T) {
	dir := t.TempDir()
	if err := WriteSchemas(dir); err != nil {
		t.Fatalf("WriteSchemas() error = %v", err)
	}

	for _, name := range []string{"cv.schema.json", "variant.schema.json"} {
		checkedInPath := filepath.Join("..", "..", "schema", name)
		checkedIn, err := os.ReadFile(checkedInPath)
		if os.IsNotExist(err) {
			t.Skipf("checked-in schema file %s does not exist", checkedInPath)
		}
		if err != nil {
			t.Fatalf("read checked-in %s: %v", name, err)
		}

		generated, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read generated %s: %v", name, err)
		}
		if !bytes.Equal(checkedIn, generated) {
			t.Fatalf("%s differs from generated output; run `cvx schema`", checkedInPath)
		}
	}
}
