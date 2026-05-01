package normalize

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/josesanchez/cvx/internal/cv"
)

type Snapshot struct {
	Input        string   `json:"input"`
	Variant      string   `json:"variant,omitempty"`
	Target       string   `json:"target,omitempty"`
	SectionOrder []string `json:"section_order"`
	CV           *cv.CV   `json:"cv"`
}

func Command(args []string) error {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		usage()
		return nil
	}

	fs := flag.NewFlagSet("normalize", flag.ContinueOnError)
	input := fs.String("input", "cv.yaml", "CV YAML path")
	variantPath := fs.String("variant", "", "variant YAML path")
	output := fs.String("output", "output/current.json", "normalized JSON output path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() > 0 {
		return fmt.Errorf("normalize: unexpected argument %q", fs.Arg(0))
	}

	if err := Write(*input, *variantPath, *output); err != nil {
		return err
	}
	fmt.Println(*output)
	return nil
}

func usage() {
	fmt.Print(`Usage:
  cvx normalize [--input cv.yaml] [--variant variants/name.yaml] [--output output/current.json]

Writes a canonical JSON snapshot for diffing and audit reports.
`)
}

func Write(input, variantPath, output string) error {
	snapshot, err := BuildSnapshot(input, variantPath)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	return os.WriteFile(output, data, 0o644)
}

func BuildSnapshot(input, variantPath string) (*Snapshot, error) {
	doc, err := cv.Load(input)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", input, err)
	}
	if errs := cv.Validate(doc); len(errs) > 0 {
		return nil, fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
	}

	sections := []string{"summary", "experience", "projects", "skills", "education"}
	snapshot := &Snapshot{Input: input, Variant: variantPath, SectionOrder: sections, CV: doc}
	if variantPath == "" {
		return snapshot, nil
	}

	variant, err := cv.LoadVariant(variantPath)
	if err != nil {
		return nil, fmt.Errorf("load variant %s: %w", variantPath, err)
	}
	if errs := cv.ValidateVariant(variant); len(errs) > 0 {
		return nil, fmt.Errorf("variant validation failed: %s", strings.Join(errs, "; "))
	}
	snapshot.Target = variant.Target
	snapshot.SectionOrder = append([]string(nil), variant.SectionOrder...)
	snapshot.CV = cv.ApplyVariant(doc, variant)
	return snapshot, nil
}
