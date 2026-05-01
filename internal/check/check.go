package check

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/josesanchez/cvx/internal/cv"
	"github.com/josesanchez/cvx/internal/diff"
	"github.com/josesanchez/cvx/internal/normalize"
	"github.com/josesanchez/cvx/internal/render"
)

func Command(args []string) error {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		usage()
		return nil
	}
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	input := fs.String("input", "cv.yaml", "CV YAML path")
	variant := fs.String("variant", "", "variant YAML path")
	previous := fs.String("previous", "output/previous.json", "previous normalized snapshot path")
	current := fs.String("current", "output/current.json", "current normalized snapshot path")
	renderOutput := fs.String("render-output", "output/cv.tex", "rendered TeX output path")
	diffOutput := fs.String("diff-output", "output/reports/last-diff.md", "diff markdown output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("check: unexpected argument %q", fs.Arg(0))
	}

	if err := lint(*input, *variant); err != nil {
		return err
	}
	fmt.Println("check: lint pass")

	if err := normalize.Write(*input, *variant, *current); err != nil {
		return err
	}
	fmt.Println("check: normalized", *current)

	renderArgs := []string{"--input", *input, "--output", *renderOutput}
	if *variant != "" {
		renderArgs = append(renderArgs, "--variant", *variant)
	}
	if err := render.Command(renderArgs); err != nil {
		return err
	}
	fmt.Println("check: render pass")

	if _, err := os.Stat(*previous); err == nil {
		if err := diff.Command([]string{"--from", *previous, "--to", *current, "--output", *diffOutput}); err != nil {
			return err
		}
		fmt.Println("check: diff", *diffOutput)
	} else if os.IsNotExist(err) {
		fmt.Println("check: diff skipped; previous snapshot missing")
	} else {
		return fmt.Errorf("stat previous snapshot %s: %w", *previous, err)
	}
	return nil
}

func lint(input, variantPath string) error {
	doc, err := cv.Load(input)
	if err != nil {
		return err
	}
	if errs := cv.Validate(doc); len(errs) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
	}
	if variantPath == "" {
		return nil
	}
	variant, err := cv.LoadVariant(variantPath)
	if err != nil {
		return err
	}
	if errs := cv.ValidateVariant(variant); len(errs) > 0 {
		return fmt.Errorf("variant validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

func usage() {
	fmt.Print(`Usage:
  cvx check [--input cv.yaml] [--variant variants/name.yaml]

Runs lint, normalize, render, and diff when a previous snapshot exists.
`)
}
