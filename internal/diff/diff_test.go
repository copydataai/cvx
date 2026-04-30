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

func TestMarkdownReportsProjectReplacementAtSameIndexAsRemovedAndAdded(t *testing.T) {
	from := &cv.CV{
		Projects: []cv.Project{
			{
				Name: "Old Project",
				Bullets: []string{
					"Built legacy dispatch workflows.",
				},
			},
		},
	}
	to := &cv.CV{
		Projects: []cv.Project{
			{
				Name: "New Project",
				Bullets: []string{
					"Created invoice review tools.",
				},
			},
		},
	}

	report := Markdown(from, to)
	assertContains(t, report, "### Old Project")
	assertContains(t, report, "Removed bullet: Built legacy dispatch workflows.")
	assertContains(t, report, "### New Project")
	assertContains(t, report, "Added bullet: Created invoice review tools.")
	assertNotContains(t, report, "Changed wording")
}

func TestMarkdownReportsUnrelatedSameLengthBulletChangeAsRemovedAndAdded(t *testing.T) {
	from := &cv.CV{
		Projects: []cv.Project{
			{
				Name: "Ops Tool",
				Bullets: []string{
					"Built scheduling workflows.",
				},
			},
		},
	}
	to := &cv.CV{
		Projects: []cv.Project{
			{
				Name: "Ops Tool",
				Bullets: []string{
					"Created invoice exports.",
				},
			},
		},
	}

	report := Markdown(from, to)
	assertContains(t, report, "Removed bullet: Built scheduling workflows.")
	assertContains(t, report, "Added bullet: Created invoice exports.")
	assertNotContains(t, report, "Changed wording")
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected report to contain %q; report:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("expected report not to contain %q; report:\n%s", want, got)
	}
}
