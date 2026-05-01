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

func Command(args []string) error {
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

func Write(input, variantPath, output string) error {
	doc, err := cv.Load(input)
	if err != nil {
		return fmt.Errorf("load %s: %w", input, err)
	}
	if errs := cv.Validate(doc); len(errs) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
	}

	if variantPath != "" {
		variant, err := cv.LoadVariant(variantPath)
		if err != nil {
			return fmt.Errorf("load variant %s: %w", variantPath, err)
		}
		if errs := cv.ValidateVariant(variant); len(errs) > 0 {
			return fmt.Errorf("variant validation failed: %s", strings.Join(errs, "; "))
		}
		doc = cv.ApplyVariant(doc, variant)
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	return os.WriteFile(output, data, 0o644)
}
