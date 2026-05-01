package main

import (
	"fmt"
	"os"

	"github.com/josesanchez/cvx/internal/check"
	"github.com/josesanchez/cvx/internal/cv"
	"github.com/josesanchez/cvx/internal/diff"
	"github.com/josesanchez/cvx/internal/normalize"
	"github.com/josesanchez/cvx/internal/prompt"
	"github.com/josesanchez/cvx/internal/render"
	"github.com/josesanchez/cvx/internal/schema"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "cvx:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		usage()
		return nil
	}

	switch args[0] {
	case "init":
		return cv.WriteStarterFiles(".")
	case "check":
		return check.Command(args[1:])
	case "lint":
		return cv.LintCommand(args[1:])
	case "render":
		return render.Command(args[1:])
	case "normalize":
		return normalize.Command(args[1:])
	case "diff":
		return diff.Command(args[1:])
	case "prompt":
		return prompt.Command(args[1:])
	case "schema":
		return schemaCommand(args[1:])
	case "help", "--help", "-h":
		usage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func schemaCommand(args []string) error {
	if len(args) > 0 {
		if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
			schemaUsage()
			return nil
		}
		return fmt.Errorf("schema: unknown argument %q", args[0])
	}

	if err := schema.WriteSchemas("schema"); err != nil {
		return err
	}
	fmt.Println("schema/cv.schema.json")
	fmt.Println("schema/variant.schema.json")
	fmt.Println("schema/snapshot.schema.json")
	return nil
}

func usage() {
	fmt.Print(`cvx is an agent-native CV system.

Usage:
  cvx init
  cvx check [--input cv.yaml] [--variant variants/name.yaml]
  cvx lint [--input cv.yaml] [--variant variants/name.yaml]
  cvx render [--input cv.yaml] [--variant variants/name.yaml] [--output output/cv.tex]
  cvx normalize [--input cv.yaml] [--variant variants/name.yaml] [--output output/current.json]
  cvx diff --from output/previous.json --to output/current.json [--output output/reports/last-diff.md]
  cvx prompt list|<name>
  cvx schema
`)
}

func schemaUsage() {
	fmt.Print(`Usage:
  cvx schema

Writes static JSON Schema files to:
  schema/cv.schema.json
  schema/variant.schema.json
  schema/snapshot.schema.json
`)
}
