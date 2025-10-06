# Tic-Tac-Turing

A playful Model Context Protocol (MCP) powered twist on tic-tac-toe where a human challenges an LLM champion.

## Project Structure

```
tic-tac-turing/
├── cmd/
│   └── server/
│       └── main.go           # Main application entry point
├── internal/
│   ├── mcp/
│   │   └── handler.go        # MCP server endpoint (stub)
│   └── web/
│       └── handler.go        # Web handlers for HTML and static files
├── web/
│   └── static/
│       ├── index.html        # Main website content
│       └── styles.css        # CSS styling
├── go.mod                    # Go module definition
└── README.md                 # This file
```

## Development

### Running the Server

```bash
# Build the server
go build -o bin/server ./cmd/server

# Run the server
./bin/server

# Or run directly without building
go run ./cmd/server
```

The server will start on port 8080 by default. You can override this by setting the `PORT` environment variable:

```bash
PORT=3000 go run ./cmd/server
```

### Endpoints

- `/` - Main website (serves `index.html` with dynamic modification support)
- `/styles.css` - CSS stylesheet
- `/mcp` - MCP server endpoint (stub - implement your MCP logic here)

### Adding Dynamic Content to HTML

The `internal/web/handler.go` file includes a `modifyHTML()` function that allows you to dynamically modify the HTML before serving it. This is where you'll add the high-scores section:

```go
func modifyHTML(content []byte) []byte {
    // Example: inject high scores
    // marker := []byte("<!-- HIGH_SCORES_PLACEHOLDER -->")
    // scoresHTML := generateHighScoresHTML(scores)
    // return bytes.Replace(content, marker, scoresHTML, 1)
    
    return content
}
```

To use this:
1. Add a placeholder comment in `index.html` where you want dynamic content
2. Implement the logic to generate and inject the content in `modifyHTML()`

### Implementing the MCP Server

The MCP server handler stub is located at `internal/mcp/handler.go`. Replace the stub implementation with your actual MCP server logic.

## Deployment

This project is designed to be deployed at `https://tic-tac-turing.fly.dev`.

### Environment Variables

- `PORT` - Server port (default: 8080)

## License

See project documentation for license information.
