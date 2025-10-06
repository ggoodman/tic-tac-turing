# Web Content Architecture

## Overview

The website content is authored in **Markdown** and served with **content negotiation** support, allowing clients to request either raw markdown or rendered HTML.

## File Structure

```
web/
├── content/
│   ├── embed.go       # Embeds *.md files using go:embed
│   ├── render.go      # Renders markdown to HTML at startup
│   └── home.md        # Main landing page content (markdown source)
└── static/
    ├── embed.go       # Embeds CSS files
    └── styles.css     # Stylesheet (referenced by rendered HTML)
```

## How It Works

### 1. Startup (Initialization)

When the server starts, `web.Init()` is called from `main()`:

- Reads `web/content/home.md` from the embedded filesystem
- Parses markdown using [goldmark](https://github.com/yuin/goldmark) (GitHub Flavored Markdown)
- Wraps rendered HTML in a minimal HTML document with:
  - Proper DOCTYPE, meta tags, and charset
  - Link to `/styles.css` for styling
- Stores both raw markdown and rendered HTML in memory

### 2. Request Handling (Content Negotiation)

The `web.Handler` serves requests to `/`, `/home.md`, and `/index.md` with content negotiation:

**Decision order:**
1. **Explicit `.md` suffix** (`/home.md`, `/index.md`) → serve markdown
2. **Query parameter** (`?format=md`) → serve markdown
3. **Accept header**: `text/markdown` without `text/html` → serve markdown
4. **Default** → serve HTML

**Response:**
- Markdown: `Content-Type: text/markdown; charset=utf-8`
- HTML: `Content-Type: text/html; charset=utf-8`

### 3. Example Requests

```bash
# Get HTML (default)
curl http://localhost:8080/

# Get HTML (explicit)
curl -H "Accept: text/html" http://localhost:8080/

# Get markdown via path
curl http://localhost:8080/home.md

# Get markdown via query
curl http://localhost:8080/?format=md

# Get markdown via Accept header
curl -H "Accept: text/markdown" http://localhost:8080/
```

## Design Rationale

### Why Markdown Source?

- **Readable**: Easy to edit, review, and version control
- **Portable**: Can be rendered by GitHub, documentation sites, etc.
- **Single source**: One file generates multiple output formats

### Why Render at Startup?

- **Performance**: No per-request parsing overhead
- **Fail-fast**: Invalid markdown causes startup failure (not runtime errors)
- **Simplicity**: No caching complexity

### Why Content Negotiation?

- **Developer-friendly**: Markdown is easier to read in terminals/editors
- **User-friendly**: HTML provides proper styling and navigation
- **Flexible**: Clients choose the format they prefer

## Dependencies

- **[goldmark](https://github.com/yuin/goldmark)**: Standards-compliant, extensible markdown parser
  - Supports GitHub Flavored Markdown (GFM)
  - Used by Hugo, Gitea, and other prominent Go projects
  - Small footprint, no external dependencies

## Testing

Run tests:

```bash
go test ./internal/web
```

The test suite validates:
- Content negotiation logic (all paths, headers, query params)
- Correct Content-Type headers
- Both markdown and HTML outputs contain expected content
- 404 handling for non-existent paths
