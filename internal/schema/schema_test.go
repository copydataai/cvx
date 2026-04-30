package schema

import (
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
