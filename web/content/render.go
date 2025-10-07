// Package content provides rendering of embedded markdown content to HTML.
package content

import (
	"bytes"
	"fmt"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/mermaid"
)

// Page holds both raw markdown and rendered HTML for a content page.
type Page struct {
	Raw  []byte
	HTML []byte
}

// htmlTemplate wraps rendered markdown content in a minimal HTML document.
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Tick-Tack-Turing</title>
  <meta name="description" content="Tick-Tack-Turing: a playful Model Context Protocol (MCP) powered twist on tic-tac-toe where a human challenges an LLM champion.">
	<link rel="stylesheet" href="styles.css">
	<link rel="icon" type="image/png" sizes="32x32" href="favicon-32x32.png">
	<link rel="icon" type="image/png" sizes="16x16" href="favicon-16x16.png">
	<link rel="apple-touch-icon" sizes="180x180" href="apple-touch-icon.png">
	<link rel="manifest" href="site.webmanifest">
	<link rel="shortcut icon" href="favicon.ico">
</head>
<body>
%s
</body>
</html>
`

// MustLoadHome loads and renders the home.md file at startup.
// Panics if the file is missing or markdown rendering fails (fail-fast).
func MustLoadHome() *Page {
	raw, err := Files.ReadFile("home.md")
	if err != nil {
		panic(fmt.Sprintf("failed to read home.md: %v", err))
	}

	// Configure goldmark with GitHub-flavored Markdown extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,         // GitHub Flavored Markdown
			extension.Typographer, // Smart quotes, dashes
			&mermaid.Extender{},   // Server-side mermaid rendering
			highlighting.NewHighlighting( // Syntax highlighting for fenced code blocks
				highlighting.WithStyle("github"), // Chroma style name
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),      // Use CSS classes instead of inline styles
					chromahtml.WithLineNumbers(false), // Keep it clean; enable later if desired
				),
			),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // Allow raw HTML in markdown if needed
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(raw, &buf); err != nil {
		panic(fmt.Sprintf("failed to render home.md: %v", err))
	}

	// Wrap rendered content in HTML template
	rendered := fmt.Sprintf(htmlTemplate, buf.String())

	return &Page{
		Raw:  raw,
		HTML: []byte(rendered),
	}
}
