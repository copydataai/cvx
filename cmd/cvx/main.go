package main

import (
	"fmt"
	"os"

	"github.com/josesanchez/cvx/internal/cv"
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
	case "lint":
		return cv.LintCommand(args[1:])
	case "render":
		return render.Command(args[1:])
	case "schema":
		if err := schema.WriteSchemas("schema"); err != nil {
			return err
		}
		fmt.Println("schema/cv.schema.json")
		fmt.Println("schema/variant.schema.json")
		return nil
	case "help", "--help", "-h":
		usage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usage() {
	fmt.Print(`cvx is an agent-native CV system.

Usage:
  cvx init
  cvx lint [--input cv.yaml] [--variant variants/name.yaml]
  cvx render [--input cv.yaml] [--variant variants/name.yaml] [--output output/cv.tex]
  cvx schema
`)
}
