package cv

import (
	"flag"
	"fmt"
	"strings"
)

func LintCommand(args []string) error {
	fs := flag.NewFlagSet("lint", flag.ContinueOnError)
	input := fs.String("input", "cv.yaml", "CV YAML path")
	variantPath := fs.String("variant", "", "variant YAML path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	doc, err := Load(*input)
	if err != nil {
		return err
	}
	errs := Validate(doc)
	if *variantPath != "" {
		variant, err := LoadVariant(*variantPath)
		if err != nil {
			return err
		}
		errs = append(errs, ValidateVariant(variant)...)
	}
	if len(errs) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
	}
	fmt.Println("cvx lint: pass")
	return nil
}
