package session

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/josesanchez/cvx/internal/check"
	"github.com/josesanchez/cvx/internal/normalize"
	"github.com/josesanchez/cvx/internal/policy"
)

const timeFormat = "20060102T150405Z"

func Command(args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		usage()
		return nil
	}
	switch args[0] {
	case "start":
		return start(args[1:])
	case "snapshot":
		return snapshot(args[1:])
	case "verify":
		return verify(args[1:])
	case "report":
		return sessionReport(args[1:])
	default:
		return fmt.Errorf("session: unknown subcommand %q", args[0])
	}
}

func start(args []string) error {
	fs := flag.NewFlagSet("session start", flag.ContinueOnError)
	goal := fs.String("goal", "", "session goal")
	dir := fs.String("dir", "", "session directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*goal) == "" {
		return fmt.Errorf("session start: --goal is required")
	}
	path := *dir
	if path == "" {
		path = filepath.Join(".cvx", "sessions", time.Now().UTC().Format(timeFormat))
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(path, "goal.md"), []byte(*goal+"\n"), 0o644); err != nil {
		return err
	}
	manifest := "goal: " + quote(*goal) + "\nfactual_changes: none\npresentation_changes: none\nuncertainty: none\n"
	if err := os.WriteFile(filepath.Join(path, "manifest.yaml"), []byte(manifest), 0o644); err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

func snapshot(args []string) error {
	fs := flag.NewFlagSet("session snapshot", flag.ContinueOnError)
	dir := fs.String("dir", latestSessionDir(), "session directory")
	label := fs.String("label", "before", "snapshot label")
	input := fs.String("input", "cv.yaml", "CV YAML path")
	variant := fs.String("variant", "", "variant YAML path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := os.MkdirAll(*dir, 0o755); err != nil {
		return err
	}
	out := filepath.Join(*dir, *label+".json")
	if err := normalize.Write(*input, *variant, out); err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

func verify(args []string) error {
	fs := flag.NewFlagSet("session verify", flag.ContinueOnError)
	dir := fs.String("dir", latestSessionDir(), "session directory")
	variant := fs.String("variant", "", "variant YAML path")
	approveHigh := fs.Bool("approve-high", false, "allow high-risk policy changes")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := check.Command(checkArgs(*variant, *dir)); err != nil {
		return err
	}
	before := filepath.Join(*dir, "before.json")
	after := filepath.Join(*dir, "after.json")
	if _, err := os.Stat(before); err != nil {
		return fmt.Errorf("session verify: missing before snapshot %s", before)
	}
	if _, err := os.Stat(after); err != nil {
		return fmt.Errorf("session verify: missing after snapshot %s", after)
	}
	policyReport := filepath.Join(*dir, "policy-report.json")
	policyArgs := []string{"--before", before, "--after", after, "--output", policyReport}
	if *approveHigh {
		policyArgs = append(policyArgs, "--approve-high")
	}
	return policy.Command(policyArgs)
}

func checkArgs(variant, dir string) []string {
	args := []string{"--current", filepath.Join(dir, "after.json"), "--previous", filepath.Join(dir, "before.json"), "--diff-output", filepath.Join(dir, "diff.md"), "--save-history=false"}
	if variant != "" {
		args = append(args, "--variant", variant)
	}
	return args
}

func sessionReport(args []string) error {
	fs := flag.NewFlagSet("session report", flag.ContinueOnError)
	dir := fs.String("dir", latestSessionDir(), "session directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	body := "# cvx Session Report\n\n"
	for _, name := range []string{"goal.md", "manifest.yaml", "diff.md", "policy-report.json"} {
		path := filepath.Join(*dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		body += "## " + name + "\n\n```\n" + string(data) + "\n```\n\n"
	}
	out := filepath.Join(*dir, "agent-report.md")
	if err := os.WriteFile(out, []byte(body), 0o644); err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

func latestSessionDir() string {
	entries, err := os.ReadDir(filepath.Join(".cvx", "sessions"))
	if err != nil || len(entries) == 0 {
		return filepath.Join(".cvx", "sessions", time.Now().UTC().Format(timeFormat))
	}
	latest := entries[len(entries)-1].Name()
	return filepath.Join(".cvx", "sessions", latest)
}

func quote(value string) string {
	return fmt.Sprintf("%q", value)
}

func usage() {
	fmt.Print(`Usage:
  cvx session start --goal "Tailor for YC founder engineer"
  cvx session snapshot --label before [--variant variants/name.yaml]
  cvx session snapshot --label after [--variant variants/name.yaml]
  cvx session verify [--variant variants/name.yaml] [--approve-high]
  cvx session report
`)
}
