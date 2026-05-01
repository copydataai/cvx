package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const dir = "prompts"

func Command(args []string) error {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		usage()
		return nil
	}
	if len(args) != 1 {
		return fmt.Errorf("prompt: expected exactly one prompt name")
	}
	name := strings.TrimSuffix(args[0], ".md")
	if name == "list" {
		return list()
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.TrimSpace(name) == "" {
		return fmt.Errorf("prompt: invalid prompt name %q", args[0])
	}
	path := filepath.Join(dir, name+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read prompt %s: %w", path, err)
	}
	fmt.Print(string(data))
	return nil
}

func list() error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), ".md"))
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Println(name)
	}
	return nil
}

func usage() {
	fmt.Print(`Usage:
  cvx prompt list
  cvx prompt <name>

Prints a prompt from prompts/<name>.md.
`)
}
