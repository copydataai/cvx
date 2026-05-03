package preview

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/josesanchez/cvx/internal/render"
)

type config struct {
	addr    string
	input   string
	variant string
	html    string
	once    bool
	watch   bool
	poll    time.Duration
}

type rebuildState struct {
	mu      sync.RWMutex
	lastRun time.Time
	lastErr string
}

func Command(args []string) error {
	cfg, err := parseFlags(args)
	if err != nil {
		return err
	}

	state := &rebuildState{}
	if err := rebuild(cfg, state); err != nil && cfg.once {
		return err
	}
	if cfg.once {
		return nil
	}

	if cfg.watch {
		go watch(cfg, state)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", dashboardHandler(state))
	mux.HandleFunc("/rebuild", rebuildHandler(cfg, state))
	mux.HandleFunc("/artifact/html", artifactHandler(cfg.html, "text/html; charset=utf-8"))
	mux.HandleFunc("/artifact/tex", artifactHandler("output/cv.tex", "text/plain; charset=utf-8"))
	mux.HandleFunc("/artifact/render-report", artifactHandler("output/reports/last-render.json", "application/json; charset=utf-8"))
	mux.HandleFunc("/artifact/diff-report", artifactHandler("output/reports/last-diff.md", "text/markdown; charset=utf-8"))
	mux.HandleFunc("/artifact/current-json", artifactHandler("output/current.json", "application/json; charset=utf-8"))
	mux.HandleFunc("/artifact/review-facts", artifactHandler("output/reports/review-facts.md", "text/markdown; charset=utf-8"))
	mux.HandleFunc("/artifact/review-bullets", artifactHandler("output/reports/review-bullets.md", "text/markdown; charset=utf-8"))
	mux.HandleFunc("/artifact/review-ats", artifactHandler("output/reports/review-ats.md", "text/markdown; charset=utf-8"))
	mux.HandleFunc("/artifact/review-target", artifactHandler("output/reports/review-target.md", "text/markdown; charset=utf-8"))

	fmt.Println("preview server listening on", cfg.addr)
	return http.ListenAndServe(cfg.addr, mux)
}

func parseFlags(args []string) (config, error) {
	fs := flag.NewFlagSet("preview", flag.ContinueOnError)
	cfg := config{}
	fs.StringVar(&cfg.addr, "addr", "127.0.0.1:4321", "preview server address")
	fs.StringVar(&cfg.input, "input", "cv.yaml", "CV YAML path")
	fs.StringVar(&cfg.variant, "variant", "", "variant YAML path")
	fs.StringVar(&cfg.html, "html", "output/cv.html", "HTML output path")
	fs.BoolVar(&cfg.once, "once", false, "rebuild once without serving")
	fs.BoolVar(&cfg.watch, "watch", true, "poll source files and rebuild when they change")
	fs.DurationVar(&cfg.poll, "poll", time.Second, "file watch polling interval")
	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	return cfg, nil
}

func watch(cfg config, state *rebuildState) {
	if cfg.poll <= 0 {
		cfg.poll = time.Second
	}
	ticker := time.NewTicker(cfg.poll)
	defer ticker.Stop()
	last := latestModTime(watchPaths(cfg))
	for range ticker.C {
		next := latestModTime(watchPaths(cfg))
		if next.After(last) {
			_ = rebuild(cfg, state)
			last = next
		}
	}
}

func watchPaths(cfg config) []string {
	paths := []string{cfg.input, "templates/html/minimal.html.tmpl", "templates/html/founder.html.tmpl"}
	if cfg.variant != "" {
		paths = append(paths, cfg.variant)
	}
	return paths
}

func latestModTime(paths []string) time.Time {
	var latest time.Time
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
	}
	return latest
}

func rebuild(cfg config, state *rebuildState) error {
	args := []string{"--input", cfg.input, "--format", "html", "--output", cfg.html}
	if cfg.variant != "" {
		args = append(args, "--variant", cfg.variant)
	}
	err := render.Command(args)
	state.set(err)
	return err
}

func (s *rebuildState) set(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastRun = time.Now()
	if err != nil {
		s.lastErr = err.Error()
		return
	}
	s.lastErr = ""
}

func (s *rebuildState) snapshot() (time.Time, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRun, s.lastErr
}

func dashboardHandler(state *rebuildState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		lastRun, lastErr := state.snapshot()
		data := struct {
			Status  string
			LastRun string
			LastErr string
		}{
			Status:  "ok",
			LastRun: "never",
			LastErr: lastErr,
		}
		if !lastRun.IsZero() {
			data.LastRun = lastRun.Format(time.RFC3339)
		}
		if lastErr != "" {
			data.Status = "error"
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := dashboardTemplate.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func rebuildHandler(cfg config, state *rebuildState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		_ = rebuild(cfg, state)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func artifactHandler(path, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if path == "" {
			http.NotFound(w, r)
			return
		}
		info, err := os.Stat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if info.IsDir() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", contentType)
		http.ServeFile(w, r, filepath.Clean(path))
	}
}

var dashboardTemplate = template.Must(template.New("dashboard").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>cvx preview</title>
  <style>
    body { margin: 0; font-family: ui-sans-serif, system-ui, sans-serif; background: #f6f2ea; color: #17130d; }
    header { display: flex; gap: 1rem; align-items: center; justify-content: space-between; padding: 1rem 1.25rem; border-bottom: 1px solid #ded6c8; background: #fffaf0; }
    main { display: grid; grid-template-columns: 220px 1fr; min-height: calc(100vh - 86px); }
    nav.tabs { padding: 1rem; border-right: 1px solid #ded6c8; background: #fdf7eb; display: flex; flex-direction: column; gap: .5rem; }
    nav.tabs a { color: #173f35; font-weight: 800; text-decoration: none; padding: .55rem .7rem; border: 1px solid #d8cfbf; background: #fffaf0; }
    .status { display: grid; gap: .25rem; font-size: .95rem; }
    .error { color: #8b1e16; white-space: pre-wrap; }
    .panel { padding: 1rem; }
    iframe { width: 100%; height: calc(100vh - 8rem); border: 1px solid #ded6c8; background: white; }
    .quicklinks { display: flex; flex-wrap: wrap; gap: .75rem; }
    .quicklinks a { color: #173f35; font-weight: 700; }
    @media (max-width: 760px) { main { grid-template-columns: 1fr; } nav.tabs { border-right: 0; border-bottom: 1px solid #ded6c8; } }
  </style>
</head>
<body>
  <header>
    <div class="status">
      <strong>Rebuild status: {{.Status}}</strong>
      <span>Last run: {{.LastRun}}</span>
      <span>Watching source files for changes.</span>
      {{if .LastErr}}<span class="error">Last error: {{.LastErr}}</span>{{end}}
    </div>
    <div class="quicklinks"><a href="/rebuild">Rebuild</a></div>
  </header>
  <main>
    <nav class="tabs">
      <a href="/artifact/html" target="preview-frame">HTML</a>
      <a href="/artifact/tex" target="preview-frame">TeX</a>
      <a href="/artifact/render-report" target="preview-frame">Render Report</a>
      <a href="/artifact/diff-report" target="preview-frame">Diff Report</a>
      <a href="/artifact/current-json" target="preview-frame">Snapshot</a>
      <a href="/artifact/review-facts" target="preview-frame">Facts Review</a>
      <a href="/artifact/review-bullets" target="preview-frame">Bullets Review</a>
      <a href="/artifact/review-ats" target="preview-frame">ATS Review</a>
      <a href="/artifact/review-target" target="preview-frame">Target Review</a>
    </nav>
    <section class="panel">
      <iframe name="preview-frame" title="CV preview artifact" src="/artifact/html"></iframe>
    </section>
  </main>
</body>
</html>`))
