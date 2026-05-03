package pdf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Result struct {
	Status string
	Engine string
	Output string
	Error  string
}

func Compile(texPath, pdfPath, requested string) Result {
	if requested == "none" || requested == "false" {
		return Result{Status: "skipped", Engine: "tex-only", Output: pdfPath}
	}
	engine, ok := selectEngine(requested)
	if !ok {
		return Result{Status: "skipped", Engine: "tex-only", Output: pdfPath, Error: "no supported LaTeX engine found"}
	}
	if err := os.MkdirAll(filepath.Dir(pdfPath), 0o755); err != nil {
		return Result{Status: "failed", Engine: engine, Output: pdfPath, Error: err.Error()}
	}
	if err := runEngine(engine, texPath, pdfPath); err != nil {
		return Result{Status: "failed", Engine: engine, Output: pdfPath, Error: err.Error()}
	}
	return Result{Status: "success", Engine: engine, Output: pdfPath}
}

func selectEngine(requested string) (string, bool) {
	if requested != "" && requested != "auto" {
		_, err := exec.LookPath(requested)
		return requested, err == nil
	}
	for _, engine := range []string{"tectonic", "latexmk", "pdflatex"} {
		if _, err := exec.LookPath(engine); err == nil {
			return engine, true
		}
	}
	return "", false
}

func runEngine(engine, texPath, pdfPath string) error {
	dir := filepath.Dir(texPath)
	base := filepath.Base(texPath)
	switch engine {
	case "tectonic":
		cmd := exec.Command("tectonic", "--outdir", filepath.Dir(pdfPath), texPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("tectonic failed: %s", string(out))
		}
	case "latexmk":
		cmd := exec.Command("latexmk", "-pdf", "-interaction=nonstopmode", "-halt-on-error", "-outdir="+filepath.Dir(pdfPath), texPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("latexmk failed: %s", string(out))
		}
	case "pdflatex":
		cmd := exec.Command("pdflatex", "-interaction=nonstopmode", "-halt-on-error", "-output-directory", filepath.Dir(pdfPath), base)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("pdflatex failed: %s", string(out))
		}
	default:
		return fmt.Errorf("unsupported PDF engine %q", engine)
	}
	produced := filepath.Join(filepath.Dir(pdfPath), trimExt(filepath.Base(texPath))+".pdf")
	if produced != pdfPath {
		if err := os.Rename(produced, pdfPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func trimExt(name string) string {
	ext := filepath.Ext(name)
	return name[:len(name)-len(ext)]
}
