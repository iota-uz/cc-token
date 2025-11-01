# cc-token

A production-grade command-line tool for counting tokens in files and directories using Anthropic's Claude API.

## Features

- **Single File or Directory Processing**: Count tokens for individual files or entire directory trees
- **Token Visualization**: Web-based interactive viewer, HTML export, and colored terminal visualization
- **Token Analysis**: Comprehensive token optimization analysis with recommendations (files only)
- **Smart Caching**: Local file-based cache to avoid redundant API calls
- **Cost Estimation**: Automatic cost calculation based on current Claude API pricing
- **Gitignore Support**: Respects `.gitignore` files to skip unwanted files
- **Extension Filtering**: Filter files by extension (e.g., only `.go` and `.md` files)
- **Parallel Processing**: Concurrent API requests with configurable concurrency
- **Multiple Output Formats**: Human-readable tree view or JSON output
- **Stdin Support**: Pipe content directly from stdin
- **File Size Limits**: Configurable maximum file size to avoid processing huge files
- **Production-Grade**: Proper error handling, timeouts, and verbose logging

## Installation

### Using go install (Recommended)

```bash
go install github.com/iota-uz/cc-token@latest
```

This installs the `cc-token` binary to your `$GOPATH/bin` directory.

### From Source

```bash
git clone https://github.com/iota-uz/cc-token.git
cd cc-token
go build -o cc-token .
```

## Prerequisites

You need an Anthropic API key. Get one from [console.anthropic.com](https://console.anthropic.com/).

Set it as an environment variable:

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

Add this to your `~/.bashrc`, `~/.zshrc`, or equivalent to persist across sessions.

## Usage

cc-token uses a subcommand-based interface:

```bash
cc-token <command> [flags] [arguments]
```

### Commands

| Command     | Description                           |
|-------------|---------------------------------------|
| `count`     | Count tokens in files or directories  |
| `visualize` | Visualize individual tokens in a file |
| `cache`     | Manage the token count cache          |

### Global Flags

Available for all commands:

| Flag            | Short | Type    | Default             | Description                                     |
|-----------------|-------|---------|---------------------|-------------------------------------------------|
| `--model`       | `-m`  | string  | `claude-sonnet-4-5` | Model to use for token counting                 |
| `--ext`         | `-e`  | strings | `[]`                | File extensions to include (e.g., .go,.txt,.md) |
| `--max-size`    |       | int64   | `2097152`           | Maximum file size in bytes (2MB)                |
| `--concurrency` | `-c`  | int     | `5`                 | Number of concurrent API requests               |
| `--show-cost`   |       | bool    | `true`              | Show estimated API cost                         |
| `--json`        | `-j`  | bool    | `false`             | Output results in JSON format                   |
| `--verbose`     | `-v`  | bool    | `false`             | Enable verbose output (shows cache hits)        |
| `--no-cache`    |       | bool    | `false`             | Disable caching                                 |
| `--yes`         | `-y`  | bool    | `false`             | Skip confirmation prompts (for automation)      |
| `--plain`       |       | bool    | `false`             | Use plain text output (no ANSI colors)          |
| `--output`      | `-o`  | string  | `""`                | Output file path for HTML export                |
| `--no-browser`  |       | bool    | `false`             | Skip auto-opening browser for web visualization |
| `--analyze`     |       | bool    | `false`             | Perform token optimization analysis (files only) |

## Examples

### Single File

Count tokens in a single file using the default model:

```bash
cc-token count document.txt
```

Output:

```
document.txt: 1234 tokens
Estimated cost: $0.003702
```

### Directory (Tree View)

Count tokens in all files within a directory:

```bash
cc-token count .
```

Output:

```
./
├─ main.go: 5432 tokens
├─ README.md: 987 tokens
└─ go.mod: 45 tokens
--------------------------------------------------
Total: 6464 tokens across 3 files
Estimated cost: $0.019392
```

### With Extension Filter

Count only Go and Markdown files:

```bash
cc-token count --ext .go,.md src/
```

### With Specific Model

Use Claude Opus 4.1 with the full model name:

```bash
cc-token count --model claude-opus-4-1 document.txt
```

Or use convenient aliases for the latest Claude 4.x models:

```bash
cc-token count --model haiku document.txt    # Uses claude-haiku-4-5 (fastest, cheapest)
cc-token count --model opus document.txt     # Uses claude-opus-4-1 (most capable)
cc-token count --model sonnet document.txt   # Uses claude-sonnet-4-5 (default, best balance)
```

### JSON Output

Get results in JSON format (useful for scripting):

```bash
cc-token count --json . > tokens.json
```

Output:

```json
[
  {
    "path": ".",
    "tokens": 6464,
    "type": "directory",
    "files": 3,
    "estimated_cost": 0.019392
  }
]
```

### From Stdin

Pipe content directly:

```bash
cat large-file.txt | cc-token count -
echo "Hello, Claude!" | cc-token count -
```

### Multiple Files

Process multiple files in one command:

```bash
cc-token count file1.txt file2.txt file3.txt
```

### Verbose Mode

See which files are served from cache:

```bash
cc-token count --verbose .
```

Output:

```
./
├─ main.go: 5432 tokens (cached)
├─ README.md: 987 tokens
└─ go.mod: 45 tokens (cached)
```

### Disable Caching

Force fresh API calls:

```bash
cc-token count --no-cache document.txt
```

### High Concurrency

Process large directories faster:

```bash
cc-token count --concurrency 10 ./large-project
```

### Large Files

Increase max file size to 50MB:

```bash
cc-token count --max-size 52428800 ./data
```

### Clear Cache

Remove all cached token counts:

```bash
cc-token cache clear
```

## Token Analysis

`cc-token` can analyze individual files to identify token optimization opportunities and provide actionable recommendations.

### Overview

The `--analyze` flag performs a comprehensive analysis of a single file's token usage patterns, including:

- **Efficiency Score**: Overall assessment of token usage (0-100 scale)
- **Token Density Heatmap**: Visual representation of token distribution across the file
- **Category Breakdown**: Distribution by content type (prose, code, URLs, formatting, whitespace)
- **Line-by-Line Insights**: Detailed analysis of the 25 most token-expensive lines
- **Pattern Detection**: Identifies optimization opportunities like repeated URLs, excessive whitespace, and inefficient formatting
- **Actionable Recommendations**: Prioritized suggestions with estimated token savings

### Usage

Analyze a single file to identify optimization opportunities:

```bash
cc-token count --analyze document.txt
```

**Constraints:**
- Works with **single files only** (no directories or multiple files)
- Does **not support stdin** input (use `-` argument)
- Automatically skips the cost confirmation prompt for analysis mode

### Example Output

The analysis output includes:

```
📊 Token Analysis Report
================================================================================
File: document.txt
Total Tokens: 1,234  |  Lines: 45  |  Efficiency Score: 72/100
================================================================================

📈 Token Distribution (Heatmap)
Low ░░░  Moderate ▒▒▒  High ███
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 Content Categories
  📝 Prose:       456 tokens (37%)
  💻 Code:        234 tokens (19%)
  🔗 URLs:        145 tokens (12%)
  ✨ Formatting:  198 tokens (16%)
  ⬜ Whitespace:  201 tokens (16%)

🎯 Quick Wins
  • Consolidate 12 consecutive empty lines → Save 48 tokens
  • URL appears 5 times, use link references → Save 87 tokens

💡 Recommendations
  [Priority 1] Remove excessive blank lines (Lines 5-8, 12-15, 23-26)
    Before: "...\n\n\n\n..."  →  After: "...\n..."
    Estimated savings: 35 tokens

  [Priority 2] Use reference-style links for repeated URLs (Lines 3, 7, 14, 19, 31)
    Before: "See [link](https://...)"  →  After: "See [link][1]"
    Estimated savings: 42 tokens

  [Priority 3] Remove trailing whitespace from long lines
    Before: "content    "  →  After: "content"
    Estimated savings: 18 tokens

📊 Summary
  Total potential savings: 95 tokens (7.7% reduction)
```

### Output Modes

The analysis respects the `--plain` flag for plain text output without ANSI colors or emoji:

```bash
cc-token count --analyze --plain document.txt
```

This is useful for automation and piping output to other tools.

### Use Cases

**Content Optimization:**
```bash
# Identify redundant content in documentation
cc-token count --analyze README.md
```

**Code Review:**
```bash
# Find inefficient formatting in code files
cc-token count --analyze main.go
```

**LLM Prompt Optimization:**
```bash
# Reduce token usage for API requests
cc-token count --analyze prompt.txt
```

**Batch Analysis:**
```bash
# Analyze multiple files (run separately)
for file in *.md; do
  echo "=== $file ==="
  cc-token count --analyze "$file" --plain
done
```

### Analysis Details

**Efficiency Score Calculation:**
The efficiency score considers:
- Total tokens vs. total characters
- Whitespace and empty line usage
- Average token-to-character ratio
- Deviation from optimal density

Higher scores indicate more efficient token usage.

**Pattern Detection:**
The analyzer identifies:
- Consecutive empty lines (consolidation opportunity)
- Repeated URLs and phrases (reference-style linking)
- Long lines exceeding typical width (reformatting opportunity)
- Unicode characters (potential high token cost)
- Inefficient markdown formatting

**Recommendation Prioritization:**
Recommendations are sorted by:
1. **Priority 1 (High Impact)**: Significant token savings, easy to implement
2. **Priority 2 (Medium Impact)**: Moderate savings, may require minimal changes
3. **Priority 3 (Low Impact)**: Small savings, consider for comprehensive cleanup

## Token Visualization

`cc-token` supports visualizing individual tokens using Claude's streaming API. This feature helps you understand
exactly how text is tokenized.

### Visualization Modes

#### Basic Mode

Displays tokens with colored borders in your terminal:

```bash
cc-token visualize basic document.txt
```

Output shows each token with alternating colors for easy identification:

```
Token Visualization
================================================================================
Tokens: 42    Characters: 156    Model: claude-sonnet-4-5
Estimated Cost: $0.000126
================================================================================

⎡Hello⎦⎡,⎦ ⎡world⎦⎡!⎦ ⎡This⎦ ⎡is⎦ ⎡a⎦ ⎡test⎦⎡.⎦...
```

#### Interactive Mode (Web-Based)

Launches a modern web server with an interactive UI that automatically opens in your browser:

```bash
cc-token visualize interactive document.txt
```

**Features:**

- **Modern Web UI**: Beautiful, responsive interface with dark/light theme
- **Two View Modes**: Text visualization with colored tokens + detailed table view
- **Search & Filter**: Real-time token search with match highlighting
- **Statistics Panel**: Token count, avg/max/min length analysis
- **Copy to Clipboard**: Click any token to copy it
- **Keyboard Shortcuts**: Full keyboard navigation support
- **Mobile-Friendly**: Responsive design works on all devices

**Keyboard Shortcuts:**

- `Tab` - Switch between text and table view
- `/` - Focus search box
- `Esc` - Clear search and deselect
- `?` - Show help dialog
- `t` - Toggle dark/light theme
- `↑`/`↓` or `j`/`k` - Navigate tokens
- `Ctrl+C` (in terminal) - Stop server

**Server Options:**

```bash
# Launch without auto-opening browser
cc-token visualize interactive --no-browser document.txt

# Server starts on a random available port (8080+)
# Press Ctrl+C to stop the server
```

#### HTML Export Mode

Export token visualization to a self-contained HTML file that can be opened in any browser:

```bash
cc-token visualize html --output tokens.html document.txt
```

**Features:**

- **Self-Contained**: Single HTML file with inline CSS and JavaScript
- **Portable**: Share and view offline without running a server
- **Same UI**: Identical features to interactive web mode
- **No Dependencies**: Works in any modern browser
- **Automatic Export**: No cost confirmation required

**Examples:**

```bash
# Export to HTML file
cc-token visualize html --output report.html README.md

# Export and open in browser (macOS)
cc-token visualize html --output viz.html code.py && open viz.html

# Export and open in browser (Linux)
cc-token visualize html --output viz.html code.py && xdg-open viz.html

# Export with custom model
cc-token visualize html --output tokens.html --model haiku document.txt
```

#### JSON Mode (LLM-Friendly)

Outputs structured JSON data for programmatic use and LLM consumption:

```bash
cc-token visualize json document.txt
```

Output format:

```json
{
  "content": "Hello, world!",
  "model": "claude-sonnet-4-5",
  "total_tokens": 5,
  "total_chars": 13,
  "total_bytes": 13,
  "cost": 0.000015,
  "tokens": [
    {"index": 0, "text": "Hello", "position": 0, "length": 5, "byte_size": 5},
    {"index": 1, "text": ",", "position": 5, "length": 1, "byte_size": 1},
    {"index": 2, "text": " world", "position": 6, "length": 6, "byte_size": 6},
    {"index": 3, "text": "!", "position": 12, "length": 1, "byte_size": 1}
  ]
}
```

**Benefits:**
- Machine-readable and parseable by LLMs
- Includes detailed token metadata (position, length, byte size)
- Can be piped to `jq` for filtering and analysis
- Scriptable and automatable
- No interactive confirmation (auto-skips cost warning)

#### Plain Text Mode (Pipe-Friendly)

Outputs plain text without ANSI colors, perfect for pipes and simple viewing:

```bash
cc-token visualize plain document.txt
```

Output format:

```
Token Visualization (Plain Text)
================================================================================
Tokens: 5    Characters: 13    Model: claude-sonnet-4-5
Estimated Cost: $0.000015
================================================================================

Tokenized Text:
--------------------------------------------------------------------------------
Hello|,| world|!
--------------------------------------------------------------------------------

Token Details:
--------------------------------------------------------------------------------
[0] "Hello" (pos: 0, len: 5)
[1] "," (pos: 5, len: 1)
[2] " world" (pos: 6, len: 6)
[3] "!" (pos: 12, len: 1)
--------------------------------------------------------------------------------

Total: 5 tokens
```

**Benefits:**
- No ANSI color codes (works in all environments)
- Human-readable token boundaries with pipe delimiters
- Detailed token information in structured format
- LLM-friendly plain text format
- No interactive confirmation (auto-skips cost warning)

### Cost Warning

⚠️ **Important**: Token visualization uses the streaming API which costs ~3-4x more than simple token counting because
it requires a full message generation.

Before visualization runs, you'll see a cost comparison:

```
⚠️  Token Visualization Cost Warning
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Token counting (current):  $0.000126
Token visualization:       $0.000504
Cost difference:           $0.000378 (4.0x more expensive)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Proceed with visualization? [Y/n]:
```

### Visualization Examples

**Basic visualization with Haiku (cheapest):**

```bash
cc-token visualize basic --model haiku code.py
```

**Interactive exploration:**

```bash
cc-token visualize interactive README.md
```

**JSON output for LLMs and automation:**

```bash
# Output to stdout
cc-token visualize json document.txt

# Save to file
cc-token visualize json document.txt > tokens.json

# Use with jq for filtering
cc-token visualize json document.txt | jq '.tokens[] | select(.length > 5)'
```

**Plain text output for pipes:**

```bash
cc-token visualize plain document.txt
```

**Use global --json flag:**

```bash
# Override basic mode with JSON output
cc-token visualize basic --json document.txt
```

**Skip confirmation prompt (for automation):**

```bash
# Skip cost warning with --yes flag
cc-token visualize basic --yes document.txt

# Or use shorthand -y
cc-token visualize basic -y document.txt
```

**Visualize piped content:**

```bash
echo "How are tokens split?" | cc-token visualize json -
```

### Limitations

- **Single files only**: Visualization doesn't work with directories
- **No caching**: Token boundaries aren't cached (yet)
- **Requires API key**: Uses full streaming API, not just token counting endpoint

## Caching

`cc-token` automatically caches token counts to avoid redundant API calls. The cache is stored in
`~/.cc-token/cache.json`.

**Cache Invalidation**: The cache is invalidated when:

- File content changes (detected via SHA-256 hash)
- File modification time changes

**Clear Cache**:

```bash
cc-token cache clear
```

**Disable Cache**:

```bash
cc-token count --no-cache <path>
```

## LLM and Automation Usage

`cc-token` is designed to be LLM-friendly and easily integrated into automated workflows. The JSON and plain text output modes make it ideal for use with AI agents, scripts, and pipelines.

### For LLMs and AI Agents

**JSON Output for Structured Data:**

```bash
# Get structured token data
cc-token visualize json document.txt

# Get token count in JSON format
cc-token count --json document.txt
```

The JSON output includes:
- Complete token list with positions and lengths
- Byte size information for each token
- Model information and cost estimates
- Fully parseable by LLMs and scripts

**Skip Confirmation for Non-Interactive Use:**

```bash
# Auto-skip cost warnings
cc-token visualize json --yes document.txt

# Or use shorthand
cc-token visualize json -y document.txt
```

JSON and plain text modes automatically skip confirmation prompts.

### For Scripts and Pipelines

**Example: Token Analysis Script**

```bash
#!/bin/bash
# Analyze token distribution

cc-token visualize json document.txt | jq '{
  total_tokens: .total_tokens,
  avg_length: (.tokens | map(.length) | add / length),
  long_tokens: (.tokens | map(select(.length > 10)) | length)
}'
```

**Example: CI/CD Integration**

```bash
# Check if file fits in context window
TOKENS=$(cc-token count --json main.go | jq '.[0].tokens')
if [ $TOKENS -gt 100000 ]; then
  echo "File too large for context window"
  exit 1
fi
```

**Example: Batch Processing**

```bash
# Process multiple files
for file in *.txt; do
  cc-token visualize json "$file" > "tokens_${file%.txt}.json"
done
```

### Plain Text for Human-Readable Pipes

```bash
# View token boundaries without colors
cat document.txt | cc-token visualize plain -

# Save plain text visualization
cc-token visualize plain document.txt > tokens_analysis.txt
```

### Benefits for Automation

- **No Interactive Prompts**: JSON and plain modes skip confirmation automatically
- **Machine-Readable Output**: Structured data that's easy to parse
- **Exit Codes**: Proper error handling with meaningful exit codes
- **Stdin Support**: Pipe content directly without temporary files
- **Concurrent Processing**: Built-in support for parallel API requests

## Cost Estimation

Cost estimates are based on Claude API input pricing (per 1M tokens).

### Current Models (Claude 4.x) - Recommended

| Model                 | Input (per 1M) | Output (per 1M) | Context Window | Alias    |
|-----------------------|----------------|-----------------|----------------|----------|
| **Claude Sonnet 4.5** | $3.00          | $15.00          | 200K           | `sonnet` |
| **Claude Haiku 4.5**  | $1.00          | $5.00           | 200K           | `haiku`  |
| **Claude Opus 4.1**   | $15.00         | $75.00          | 200K           | `opus`   |
| **Claude Sonnet 4**   | $3.00          | $15.00          | 200K           | -        |

### Legacy Models (Claude 3.x) - Deprecated

⚠️ **Note**: Claude 3.x models are deprecated. Use Claude 4.x models for better performance and features.

| Model                 | Input (per 1M) | Output (per 1M) | Context Window |
|-----------------------|----------------|-----------------|----------------|
| **Claude Haiku 3.5**  | $0.80          | $4.00           | 200K           |
| **Claude Sonnet 3.7** | $3.00          | $15.00          | 200K           |
| **Claude Sonnet 3.5** | $3.00          | $15.00          | 200K           |
| **Claude Opus 3**     | $15.00         | $75.00          | 200K           |
| **Claude Haiku 3**    | $0.25          | $1.25           | 200K           |

**Notes**:

- `cc-token` only counts **input tokens** (the content you're analyzing)
- Output pricing is shown for reference but not calculated by this tool
- Pricing as of 2025-11-01 - check [Anthropic's pricing page](https://www.anthropic.com/pricing) for latest rates
- Disable cost estimation with `-show-cost=false`

## Gitignore Support

When processing directories, `cc-token` automatically respects `.gitignore` files in the root directory being scanned.
This means:

- `node_modules/`, `.git/`, and other ignored directories are skipped
- Ignored file patterns are excluded
- Saves API costs by not processing unnecessary files

**Important Notes**:

- Only the `.gitignore` file in the root directory being scanned is used
- Nested `.gitignore` files in subdirectories are ignored
- `.git/` directory is always ignored (even without .gitignore)

## Supported Models

All Claude models are supported. The tool accepts multiple naming formats for flexibility.

### Current Models (Claude 4.x) - Recommended

- `claude-sonnet-4-5` (default) - Best balance of performance and cost
- `claude-haiku-4-5` - Fastest, most cost-effective
- `claude-opus-4-1` - Most capable, highest cost
- `claude-sonnet-4` - Previous generation Sonnet

### Legacy Models (Claude 3.x) - Deprecated

⚠️ **These models are deprecated. Use Claude 4.x models for better performance.**

- `claude-haiku-3-5` - Previous generation Haiku
- `claude-sonnet-3-7` - Previous generation Sonnet
- `claude-sonnet-3-5` - Older Sonnet variant
- `claude-opus-3` - Previous generation Opus
- `claude-haiku-3` - Oldest Haiku version

### Model Aliases

For convenience, you can use short aliases that automatically resolve to the latest Claude 4.x models:

| Alias    | Resolves To         | Description                                          |
|----------|---------------------|------------------------------------------------------|
| `sonnet` | `claude-sonnet-4-5` | Latest Sonnet - best balance of performance and cost |
| `haiku`  | `claude-haiku-4-5`  | Latest Haiku - fastest and most cost-effective       |
| `opus`   | `claude-opus-4-1`   | Latest Opus - most capable model                     |

**Examples:**

```bash
cc-token -model haiku .          # Use Haiku 4.5 (fast and cheap)
cc-token -model opus file.txt    # Use Opus 4.1 (most capable)
cc-token -model sonnet dir/      # Use Sonnet 4.5 (default)
```

### Alternate Naming Formats

The tool supports multiple naming conventions for full model names:

- Hyphen format: `claude-sonnet-4-5`, `claude-haiku-3-5`
- Dot format: `claude-sonnet-4.5`, `claude-haiku-3.5`
- Prefix format: `claude-4-sonnet`, `claude-3-5-haiku`

**Note**: Newer models will work automatically as Anthropic releases them, defaulting to Sonnet 4.5 pricing if not in
the pricing table.

## Performance Tips

### For Large Projects

1. **Use Extension Filtering**: Only count relevant files
   ```bash
   cc-token count --ext .go,.py src/
   ```

2. **Increase Concurrency**: Process more files in parallel
   ```bash
   cc-token count --concurrency 10 .
   ```

3. **Leverage Cache**: Run twice - second run will be instant
   ```bash
   cc-token count .        # First run: ~30s
   cc-token count .        # Second run: ~0.1s
   ```

4. **Use Gitignore**: Ensure `.gitignore` is up to date to skip build artifacts

### For Single Files

- Cache is especially useful for repeated checks during development
- Use `--verbose` to confirm cache hits

## Troubleshooting

### "ANTHROPIC_API_KEY environment variable is not set"

Set your API key:

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

### "API returned status 401"

Your API key is invalid or expired. Check:

1. Key is correctly set: `echo $ANTHROPIC_API_KEY`
2. Key has correct permissions at [console.anthropic.com](https://console.anthropic.com/)

### "API returned status 429"

You've hit the rate limit. Solutions:

1. Reduce concurrency: `--concurrency 3`
2. Add delays between runs
3. Contact Anthropic to increase limits

### "file too large"

Increase max size:

```bash
cc-token count --max-size 52428800 large-file.txt  # 50MB
```

### Cache Issues

Clear the cache:

```bash
cc-token cache clear
```

### Network Timeouts

The default timeout is 30 seconds. For slow connections, this is currently hardcoded but can be modified in the source.

## Architecture

`cc-token` is designed as a simple, maintainable CLI tool:

- **Single File**: All code in `main.go` (~1200 lines)
- **Minimal Dependencies**: Core counting uses stdlib only; visualization uses fatih/color, charm/lipgloss, and
  charm/bubbletea
- **Simple Installation**: `go install` just works

## Development

### Live Reload with Air

For rapid UI iteration during development, use [Air](https://github.com/cosmtrek/air) for automatic rebuilding:

**Installation:**

```bash
# Install Air globally
go install github.com/cosmtrek/air@latest
```

**Usage:**

```bash
# Start Air with the default configuration
air

# Air will:
# 1. Build the project to tmp/main
# 2. Start the visualizer in interactive mode
# 3. Watch for changes to Go, HTML, CSS, and JS files
# 4. Automatically rebuild and restart on file changes
```

**Development Workflow:**

1. Run `air` in your terminal
2. Open `http://localhost:8080` (or the port shown) in your browser
3. Edit HTML, CSS, or JS files in `internal/server/` or `internal/visualizer/`
4. Air rebuilds automatically in ~1-2 seconds
5. Refresh your browser to see changes
6. Press `Ctrl+C` to stop

**What's Watched:**

- All `.go` files (including `cmd/` and `internal/` packages)
- HTML templates (`internal/server/templates/*.html`, `internal/visualizer/templates/*.html`)
- CSS files (`internal/server/static/style.css`)
- JavaScript files (`internal/server/static/app.js`)

**Why Rebuild for HTML/CSS/JS?**

The web server uses `go:embed` to embed static assets into the binary. Changes to these files require a full rebuild
to be reflected in the running server. Air handles this automatically.

**Configuration:**

The project includes a `.air.toml` configuration file with optimized settings. You can customize it by editing `.air.toml`.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

**Testing Changes**:

```bash
go build .
./cc-token test-file.txt
```

## Version

Current version: **1.0.0**

Check version:

```bash
cc-token -version
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Links

- [Anthropic API Documentation](https://docs.anthropic.com/)
- [Token Counting API](https://docs.anthropic.com/en/docs/build-with-claude/token-counting)
- [Claude Model Pricing](https://www.anthropic.com/pricing)
- [GitHub Repository](https://github.com/iota-uz/cc-token)

## FAQ

**Q: Does this count output tokens?**
A: No, only input tokens. Output tokens depend on Claude's response, which this tool doesn't generate.

**Q: Can I count tokens for multiple models at once?**
A: No, but you can run the command multiple times with different `-model` flags. Results will be cached, so only the
first run hits the API.

**Q: Does the cache work across different models?**
A: No, cache is model-specific. Changing models will result in new API calls.

**Q: How accurate is the cost estimation?**
A: Very accurate for input tokens. Prices are current as of November 2025 but may change.
Check [Anthropic's pricing page](https://www.anthropic.com/pricing) for latest rates. Note: Cost estimation uses input
token pricing only.

**Q: Can I use this in CI/CD pipelines?**
A: Yes! Use `--json` for structured output and check the token count programmatically.

Example:

```bash
TOKENS=$(cc-token count --json main.go | jq '.[0].tokens')
if [ $TOKENS -gt 100000 ]; then
  echo "File too large for context window"
  exit 1
fi
```

**Q: What happens if a file can't be read?**
A: The error is reported but processing continues for other files. Use `--verbose` to see all errors.

**Q: Can I exclude specific files or directories?**
A: Yes, add them to `.gitignore` in the target directory. Alternatively, use `--ext` to include only specific
extensions.
