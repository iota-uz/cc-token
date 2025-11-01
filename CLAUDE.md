# cc-token - Claude Code Instructions

## Project Overview

`cc-token` is a simple, single-file CLI tool for counting tokens in text files using Anthropic's Claude API. The tool is designed to be lightweight, easy to install, and straightforward to maintain.

## Architecture

**Type:** Standalone CLI tool (installable via `go install`)

**Structure:**
- `main.go` - Single main file with all logic (at root for simple tools)
- `go.mod` - Go module definition
- `README.md` - User documentation
- `LICENSE` - MIT License
- `.gitignore` - Git ignore rules

**Dependencies:** Standard library only (no external dependencies)

## Development Guidelines

### Code Style

- Keep it simple - this is a utility tool, not a complex application
- Use standard library only - avoid external dependencies for this simple use case
- Panic on errors is acceptable for a CLI tool (current pattern)
- No need for extensive error handling or recovery

### Making Changes

When making changes to this tool:

1. **Test the change locally:**
   ```bash
   cd /path/to/cc-token
   go build .
   ./cc-token path/to/test-file.txt
   ```

2. **Verify installation works:**
   ```bash
   go install github.com/iota-uz/cc-token@latest
   cc-token --help
   ```

3. **Update README.md** if adding new flags or changing behavior

### API Integration

The tool integrates with Anthropic's `/v1/messages/count_tokens` endpoint:
- Endpoint: `https://api.anthropic.com/v1/messages/count_tokens`
- Required header: `x-api-key` (from `ANTHROPIC_API_KEY` env var)
- Required header: `anthropic-version: 2023-06-01`
- Body format: `{"model": "...", "messages": [{"role": "user", "content": "..."}]}`

**Reference:** https://docs.anthropic.com/en/docs/build-with-claude/token-counting

### Future Enhancement Ideas

If expanding functionality, consider:
- [ ] Add `-version` flag
- [ ] Better error messages (instead of panic)
- [ ] Support reading from stdin (`cat file.txt | cc-token`)
- [ ] Support multiple files in one invocation
- [ ] Add config file support (`~/.cc-token.yaml`)
- [ ] Add validation for `ANTHROPIC_API_KEY` with helpful error message
- [ ] Add progress indicator for very large files
- [ ] Add `-json` flag for structured output
- [ ] Support batch processing

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

# Run the tool
go run . test.txt

# Expected output:
# test.txt: X tokens
```

## Maintenance Notes

- This is a simple tool - keep changes minimal
- Prioritize simplicity over features
- Standard library only - no external dependencies
- Breaking changes should bump major version
