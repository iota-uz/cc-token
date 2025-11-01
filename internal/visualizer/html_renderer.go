package visualizer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"

	"github.com/pkg/browser"
)

//go:embed templates/static.html
var htmlTemplate embed.FS

const (
	htmlFilePerm = 0644 // File permission for exported HTML files
)

// HTMLRenderer exports token visualization to a static HTML file
type HTMLRenderer struct {
	OutputFile  string // Path to save HTML file
	OpenBrowser bool   // Whether to open browser after export
}

// Render generates and saves a self-contained HTML file
func (r *HTMLRenderer) Render(result *Result) error {
	if result == nil {
		return fmt.Errorf("result is nil")
	}

	// Parse embedded template
	tmpl, err := template.New("static.html").Funcs(template.FuncMap{
		"colorIndex": func(i int) int {
			return i % 6
		},
	}).ParseFS(htmlTemplate, "templates/static.html")
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, result); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Write to file
	if err := os.WriteFile(r.OutputFile, buf.Bytes(), htmlFilePerm); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ HTML visualization exported to: %s\n", r.OutputFile)
	fmt.Fprintf(os.Stderr, "✓ File size: %d bytes\n", buf.Len())

	// Open in browser if requested
	if r.OpenBrowser {
		fileURL := "file://" + r.OutputFile
		if err := browser.OpenURL(fileURL); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Failed to open browser: %v\n", err)
			fmt.Fprintf(os.Stderr, "   Please open manually: %s\n", r.OutputFile)
		} else {
			fmt.Fprintf(os.Stderr, "✓ Opened in browser\n")
		}
	}

	return nil
}
