package diff

import (
	"strings"
	"testing"

	"github.com/josesanchez/cvx/internal/cv"
)

func TestMarkdownReportsAddedRemovedAndChangedBullets(t *testing.T) {
	from := &cv.CV{
		Summary: "Built products for operators.",
		Experience: []cv.Experience{
			{
				Company: "Acme",
				Title:   "Engineer",
				Bullets: []string{
					"Old bullet about platform work.",
				},
			},
		},
		Projects: []cv.Project{
			{
				Name: "Ops Tool",
				Bullets: []string{
					"Removed bullet about internal tooling.",
				},
			},
		},
	}
	to := &cv.CV{
		Summary: "Built products for service operators.",
		Experience: []cv.Experience{
			{
				Company: "Acme",
				Title:   "Engineer",
				Bullets: []string{
					"New bullet about platform work.",
				},
			},
		},
		Projects: []cv.Project{
			{
				Name: "Ops Tool",
				Bullets: []string{
					"Added bullet about external workflow tooling.",
					"Added second project bullet.",
				},
			},
		},
	}

	report := Markdown(from, to)
	for _, want := range []string{
		"# CV Diff Report",
		"## Changed Top-Level Sections",
		"- Summary",
		"Changed wording",
		"Old bullet",
		"New bullet",
		"Removed bullet",
		"Added bullet",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("expected report to contain %q; report:\n%s", want, report)
		}
	}
}
