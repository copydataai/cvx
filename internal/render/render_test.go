package render

import (
	"strings"
	"testing"

	"github.com/josesanchez/cvx/internal/cv"
)

func TestApplyProjectFilterIncludesOnlyNamedProjects(t *testing.T) {
	projects := []cv.Project{{Name: "Keep"}, {Name: "Drop"}}
	variant := &cv.Variant{IncludeProjects: []string{"Keep"}}
	filtered := applyProjectFilter(projects, variant)
	if len(filtered) != 1 || filtered[0].Name != "Keep" {
		t.Fatalf("unexpected filtered projects: %#v", filtered)
	}
}

func TestRenderTeXUsesVariantSectionOrder(t *testing.T) {
	doc := &cv.CV{
		Name:    "Person",
		Summary: "Summary",
		Skills:  []string{"Go"},
		Projects: []cv.Project{{
			Name:        "Project",
			Description: "Description",
			Bullets:     []string{"Built it"},
		}},
	}
	tex := renderTeX(doc, []string{"skills", "summary", "projects"})
	skillsIndex := strings.Index(tex, "\\section*{Skills}")
	summaryIndex := strings.Index(tex, "\\section*{Summary}")
	if skillsIndex == -1 || summaryIndex == -1 || skillsIndex > summaryIndex {
		t.Fatalf("expected skills before summary, got:\n%s", tex)
	}
}

func TestRenderTeXEscapesLatex(t *testing.T) {
	doc := &cv.CV{Name: "A&B", Summary: "Uses 100%"}
	tex := renderTeX(doc, []string{"summary"})
	if !strings.Contains(tex, "A\\&B") || !strings.Contains(tex, "100\\%") {
		t.Fatalf("expected escaped latex, got:\n%s", tex)
	}
}
