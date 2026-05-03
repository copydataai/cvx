package pdf

import "testing"

func TestCompileNoneSkips(t *testing.T) {
	result := Compile("cv.tex", "cv.pdf", "none")
	if result.Status != "skipped" || result.Engine != "tex-only" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
