# cc-token - Claude Code Instructions

## Project Overview

`cc-token` is a production-grade CLI tool for counting tokens in text files using Anthropic's Claude API. Built with Cobra, it features a modular architecture with proper package separation for maintainability and extensibility.

## Architecture

**Type:** Standalone CLI tool (installable via `go install`)

**Structure:**
```
cc-token/
├── cmd/                    # Cobra command layer
│   ├── root.go            # Root command and global flags
│   ├── count.go           # Count subcommand
│   ├── visualize.go       # Visualize subcommand
│   └── cache.go           # Cache management subcommand
├── internal/              # Internal packages
│   ├── api/               # Anthropic API client
│   ├── cache/             # Token count caching
│   ├── config/            # Configuration structs
│   ├── pricing/           # Model pricing and aliases
│   ├── processor/         # File/directory processing
│   ├── visualizer/        # Token visualization (renderer pattern)
│   │   ├── renderer.go         # Renderer interface and selection logic
│   │   ├── json_renderer.go    # JSON output (LLM-friendly)
│   │   ├── plain_renderer.go   # Plain text output (pipe-friendly)
│   │   ├── basic_renderer.go   # Colored terminal output
│   │   ├── web_renderer.go     # Web-based interactive viewer
│   │   ├── html_renderer.go    # Static HTML export
│   │   └── templates/          # HTML templates for static export
│   ├── server/            # HTTP server for web visualization
│   │   ├── server.go           # Web server implementation
│   │   ├── templates/          # HTML templates
│   │   └── static/             # CSS and JavaScript assets
│   └── output/            # Output formatting
├── main.go                # Entry point (~15 lines)
├── go.mod                 # Go module definition
├── README.md              # User documentation
├── LICENSE                # MIT License
└── .gitignore             # Git ignore rules
```

**Dependencies:**
- CLI Framework: `github.com/spf13/cobra`
- Visualization: `github.com/fatih/color` (terminal colors)
- Browser: `github.com/pkg/browser` (cross-platform browser launching)
- Web Server: `net/http`, `html/template`, `embed` (stdlib)

## Development Guidelines

### Code Style

- Modular architecture with proper package separation
- Each package has a single, clear responsibility
- Use Cobra for professional CLI interface
- Return errors instead of panicking (proper error handling at CLI boundary)
- Keep packages focused and testable
- Use interface-based design for extensibility (e.g., Renderer interface)

### Renderer Pattern

The visualizer package uses a **Renderer pattern** for flexible output formatting:

**Interface Definition:**
```go
type Renderer interface {
    Render(result *Result) error
}
```

**Implementations:**
1. **JSONRenderer** - Machine-readable JSON output for LLMs and automation
2. **PlainRenderer** - Plain text output without ANSI colors (pipe-friendly)
3. **BasicRenderer** - Colored terminal output with token boundaries
4. **WebRenderer** - Web-based interactive viewer (launches HTTP server)
5. **HTMLRenderer** - Static HTML file export (self-contained, portable)

**Renderer Selection Logic:**
- Priority: `--json` flag > `--plain` flag > specified mode
- Auto-skip confirmation for non-interactive modes (JSON, plain, HTML)
- WebRenderer launches server on random available port (8080+)
- HTMLRenderer requires `--output` flag to specify file path
- Located in `internal/visualizer/renderer.go`

**Web Server Architecture:**
- Embedded HTML/CSS/JS templates using `go:embed`
- Graceful shutdown with Ctrl+C signal handling
- Auto-opens browser unless `--no-browser` flag is set
- Serves visualization at `http://localhost:<port>`
- No external dependencies - fully self-contained

**Benefits:**
- Easy to add new output formats
- Clean separation of data processing and rendering
- Testable components
- Consistent interface across all renderers

### Making Changes

When making changes to this tool:

1. **Test the change locally:**
   ```bash
   cd /path/to/cc-token
   go build .
   ./cc-token count path/to/test-file.txt
   ```

2. **Verify installation works:**
   ```bash
   go install github.com/iota-uz/cc-token@latest
   cc-token --help
   ```

3. **Update README.md** if adding new commands, flags, or changing behavior

4. **Follow the package structure** - put new features in the appropriate internal package

### API Integration

The tool integrates with two Anthropic API endpoints:

**1. Token Counting (default mode):**
- Endpoint: `https://api.anthropic.com/v1/messages/count_tokens`
- Required headers: `x-api-key`, `anthropic-version: 2023-06-01`
- Body format: `{"model": "...", "messages": [{"role": "user", "content": "..."}]}`
- Reference: https://docs.anthropic.com/en/docs/build-with-claude/token-counting

**2. Streaming API (visualization mode):**
- Endpoint: `https://api.anthropic.com/v1/messages`
- Required headers: `x-api-key`, `anthropic-version: 2023-06-01`
- Body format: `{"model": "...", "messages": [...], "max_tokens": 8192, "stream": true}`
- Purpose: Extract individual token boundaries by parsing SSE stream
- Note: Costs ~3-4x more than token counting (input + output tokens)

### Future Enhancement Ideas

If expanding functionality, consider:

**Completed:**
- [x] Add `-version` flag
- [x] Support reading from stdin (`cat file.txt | cc-token`)
- [x] Support multiple files in one invocation
- [x] Add validation for `ANTHROPIC_API_KEY` with helpful error message
- [x] Add `-json` flag for structured output
- [x] Token visualization (basic and interactive modes)
- [x] LLM-friendly JSON output for token visualization
- [x] Plain text output mode (pipe-friendly, no ANSI colors)
- [x] `--yes` flag to skip confirmation prompts (automation support)
- [x] `--plain` flag for plain text output
- [x] Renderer pattern for extensible output formatting
- [x] Auto-skip confirmation for non-interactive modes
- [x] Web-based interactive visualization with HTTP server
- [x] HTML export for portable token visualization
- [x] Search and filter tokens in web UI
- [x] Statistics panel showing token metrics
- [x] Dark/light theme toggle
- [x] Copy token to clipboard functionality
- [x] `--output` flag for HTML export
- [x] `--no-browser` flag to disable auto-open

### Release Process

When ready to release a new version:

```bash
# Tag the release
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Users can then install with:
# go install github.com/iota-uz/cc-token@v1.0.0
# or
# go install github.com/iota-uz/cc-token@latest
```

## Environment Requirements

- **Go Version:** 1.24+ (specified in go.mod)
- **Runtime:** Requires `ANTHROPIC_API_KEY` environment variable
- **Network:** Requires internet access to reach Anthropic API

## Testing

Currently manual testing only. To test:

```bash
# Create a test file
echo "Hello, world! This is a test." > test.txt

# Set API key
export ANTHROPIC_API_KEY="your-key-here"

# Run the tool with different commands
go run . count test.txt
go run . visualize basic test.txt
go run . visualize interactive test.txt  # Opens browser with web UI
go run . visualize interactive --no-browser test.txt  # Server without auto-open
go run . visualize html --output test-viz.html test.txt  # Export to HTML
go run . visualize json test.txt
go run . visualize plain test.txt
go run . visualize basic --json test.txt
go run . visualize basic --yes test.txt
go run . cache clear

# Test with stdin
echo "Test input" | go run . visualize json -

# Expected outputs:
# count: test.txt: X tokens, Estimated cost: $X.XXXXXX
# basic: Colored token visualization with brackets
# interactive: Launches web server at http://localhost:XXXX
# html: Creates HTML file with embedded visualization
# json: Structured JSON with tokens array
# plain: Plain text with pipe delimiters
```

## Maintenance Notes

- Modular architecture allows for independent package updates
- Add new features in appropriate internal packages
- New commands should be added in cmd/ as separate files
- Use Cobra's built-in help and validation features
- New renderers should implement the Renderer interface
- Version 1.1.0 added LLM-friendly output formats, renderer pattern, and web-based visualization
- Web server uses embedded templates (go:embed) for self-contained binary
- HTML export creates portable, self-contained files
- Breaking changes should bump major version
