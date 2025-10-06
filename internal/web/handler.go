package web

import (
	"log"
	"net/http"
	"strings"

	"github.com/ggoodman/tic-tac-turing/web/content"
	"github.com/ggoodman/tic-tac-turing/web/static"
)

var homePage *content.Page

// Init loads and renders content at startup. Must be called before serving requests.
func Init() {
	homePage = content.MustLoadHome()
}

// Handler serves the home page with content negotiation.
// Supports:
//   - Explicit .md suffix or ?format=md → serves raw markdown (text/markdown)
//   - Accept: text/markdown (without higher-priority text/html) → serves raw markdown
//   - Default → serves rendered HTML (text/html)
//
// The markdown source is embedded at web/content/home.md and rendered once at startup.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Only handle root path
	if r.URL.Path != "/" && r.URL.Path != "/home.md" && r.URL.Path != "/index.md" {
		http.NotFound(w, r)
		return
	}

	serveMarkdown := false

	// 1. Explicit markdown paths
	if strings.HasSuffix(r.URL.Path, ".md") {
		serveMarkdown = true
	}

	// 2. Query param ?format=md
	if r.URL.Query().Get("format") == "md" {
		serveMarkdown = true
	}

	// 3. Accept header negotiation (if not already decided)
	if !serveMarkdown {
		accept := r.Header.Get("Accept")
		if accept != "" {
			// Simple heuristic: if text/markdown appears and text/html doesn't, serve markdown
			hasMarkdown := strings.Contains(accept, "text/markdown")
			hasHTML := strings.Contains(accept, "text/html") || strings.Contains(accept, "*/*")
			if hasMarkdown && !hasHTML {
				serveMarkdown = true
			}
		}
	}

	if serveMarkdown {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(homePage.Raw)
	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(homePage.HTML)
	}
}

// StylesHandler serves the CSS file
func StylesHandler(w http.ResponseWriter, r *http.Request) {
	cssContent, err := static.Files.ReadFile("styles.css")
	if err != nil {
		log.Printf("Error reading CSS file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(cssContent)
}

// Example of how you might inject content in the future:
// func injectHighScores(html []byte, scores []Score) []byte {
//     // Find a marker in your HTML like <!-- HIGH_SCORES_PLACEHOLDER -->
//     // and replace it with dynamically generated HTML
//     marker := []byte("<!-- HIGH_SCORES_PLACEHOLDER -->")
//     scoresHTML := generateHighScoresHTML(scores)
//     return bytes.Replace(html, marker, scoresHTML, 1)
// }
