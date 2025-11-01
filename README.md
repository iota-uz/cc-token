# cc-token

A production-grade command-line tool for counting tokens in files and directories using Anthropic's Claude API.

## Features

- **Single File or Directory Processing**: Count tokens for individual files or entire directory trees
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

```bash
cc-token [flags] <path-to-file-or-directory>
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-model` | string | `claude-sonnet-4-5` | Model to use for token counting |
| `-ext` | string | `""` | Comma-separated file extensions (e.g., `.go,.txt,.md`) |
| `-max-size` | int64 | `10485760` | Maximum file size in bytes (10MB) |
| `-concurrency` | int | `5` | Number of concurrent API requests |
| `-show-cost` | bool | `true` | Show estimated API cost |
| `-json` | bool | `false` | Output results in JSON format |
| `-verbose` | bool | `false` | Enable verbose output (shows cache hits) |
| `-no-cache` | bool | `false` | Disable caching |
| `-clear-cache` | bool | `false` | Clear the cache and exit |
| `-version` | bool | `false` | Print version information |

## Examples

### Single File

Count tokens in a single file using the default model:

```bash
cc-token document.txt
```

Output:
```
document.txt: 1234 tokens
Estimated cost: $0.003702
```

### Directory (Tree View)

Count tokens in all files within a directory:

```bash
cc-token .
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
cc-token -ext .go,.md src/
```

### With Specific Model

Use Claude Opus 4.1 with the full model name:

```bash
cc-token -model claude-opus-4-1 document.txt
```

Or use convenient aliases for the latest Claude 4.x models:

```bash
cc-token -model haiku document.txt    # Uses claude-haiku-4-5 (fastest, cheapest)
cc-token -model opus document.txt     # Uses claude-opus-4-1 (most capable)
cc-token -model sonnet document.txt   # Uses claude-sonnet-4-5 (default, best balance)
```

### JSON Output

Get results in JSON format (useful for scripting):

```bash
cc-token -json . > tokens.json
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
cat large-file.txt | cc-token -
echo "Hello, Claude!" | cc-token -
```

### Multiple Files

Process multiple files in one command:

```bash
cc-token file1.txt file2.txt file3.txt
```

### Verbose Mode

See which files are served from cache:

```bash
cc-token -verbose .
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
cc-token -no-cache document.txt
```

### High Concurrency

Process large directories faster:

```bash
cc-token -concurrency 10 ./large-project
```

### Large Files

Increase max file size to 50MB:

```bash
cc-token -max-size 52428800 ./data
```

## Caching

`cc-token` automatically caches token counts to avoid redundant API calls. The cache is stored in `~/.cc-token/cache.json`.

**Cache Invalidation**: The cache is invalidated when:
- File content changes (detected via SHA-256 hash)
- File modification time changes

**Clear Cache**:
```bash
cc-token -clear-cache
```

**Disable Cache**:
```bash
cc-token -no-cache <path>
```

## Cost Estimation

Cost estimates are based on Claude API input pricing (per 1M tokens).

### Current Models (Claude 4.x) - Recommended

| Model | Input (per 1M) | Output (per 1M) | Context Window | Alias |
|-------|----------------|-----------------|----------------|-------|
| **Claude Sonnet 4.5** | $3.00 | $15.00 | 200K | `sonnet` |
| **Claude Haiku 4.5** | $1.00 | $5.00 | 200K | `haiku` |
| **Claude Opus 4.1** | $15.00 | $75.00 | 200K | `opus` |
| **Claude Sonnet 4** | $3.00 | $15.00 | 200K | - |

### Legacy Models (Claude 3.x) - Deprecated

⚠️ **Note**: Claude 3.x models are deprecated. Use Claude 4.x models for better performance and features.

| Model | Input (per 1M) | Output (per 1M) | Context Window |
|-------|----------------|-----------------|----------------|
| **Claude Haiku 3.5** | $0.80 | $4.00 | 200K |
| **Claude Sonnet 3.7** | $3.00 | $15.00 | 200K |
| **Claude Sonnet 3.5** | $3.00 | $15.00 | 200K |
| **Claude Opus 3** | $15.00 | $75.00 | 200K |
| **Claude Haiku 3** | $0.25 | $1.25 | 200K |

**Notes**:
- `cc-token` only counts **input tokens** (the content you're analyzing)
- Output pricing is shown for reference but not calculated by this tool
- Pricing as of 2025-11-01 - check [Anthropic's pricing page](https://www.anthropic.com/pricing) for latest rates
- Disable cost estimation with `-show-cost=false`

## Gitignore Support

When processing directories, `cc-token` automatically respects `.gitignore` files in the root directory being scanned. This means:

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

| Alias | Resolves To | Description |
|-------|-------------|-------------|
| `sonnet` | `claude-sonnet-4-5` | Latest Sonnet - best balance of performance and cost |
| `haiku` | `claude-haiku-4-5` | Latest Haiku - fastest and most cost-effective |
| `opus` | `claude-opus-4-1` | Latest Opus - most capable model |

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

**Note**: Newer models will work automatically as Anthropic releases them, defaulting to Sonnet 4.5 pricing if not in the pricing table.

## Performance Tips

### For Large Projects

1. **Use Extension Filtering**: Only count relevant files
   ```bash
   cc-token -ext .go,.py src/
   ```

2. **Increase Concurrency**: Process more files in parallel
   ```bash
   cc-token -concurrency 10 .
   ```

3. **Leverage Cache**: Run twice - second run will be instant
   ```bash
   cc-token .        # First run: ~30s
   cc-token .        # Second run: ~0.1s
   ```

4. **Use Gitignore**: Ensure `.gitignore` is up to date to skip build artifacts

### For Single Files

- Cache is especially useful for repeated checks during development
- Use `-verbose` to confirm cache hits

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
1. Reduce concurrency: `-concurrency 3`
2. Add delays between runs
3. Contact Anthropic to increase limits

### "file too large"

Increase max size:
```bash
cc-token -max-size 52428800 large-file.txt  # 50MB
```

### Cache Issues

Clear the cache:
```bash
cc-token -clear-cache
```

### Network Timeouts

The default timeout is 30 seconds. For slow connections, this is currently hardcoded but can be modified in the source.

## Architecture

`cc-token` is designed as a simple, maintainable CLI tool:

- **Single File**: All code in `main.go` (~800 lines)
- **Stdlib Only**: No external dependencies
- **Simple Installation**: `go install` just works

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
A: No, but you can run the command multiple times with different `-model` flags. Results will be cached, so only the first run hits the API.

**Q: Does the cache work across different models?**
A: No, cache is model-specific. Changing models will result in new API calls.

**Q: How accurate is the cost estimation?**
A: Very accurate for input tokens. Prices are current as of November 2025 but may change. Check [Anthropic's pricing page](https://www.anthropic.com/pricing) for latest rates. Note: Cost estimation uses input token pricing only.

**Q: Can I use this in CI/CD pipelines?**
A: Yes! Use `-json` for structured output and check the token count programmatically.

Example:
```bash
TOKENS=$(cc-token -json main.go | jq '.[0].tokens')
if [ $TOKENS -gt 100000 ]; then
  echo "File too large for context window"
  exit 1
fi
```

**Q: What happens if a file can't be read?**
A: The error is reported but processing continues for other files. Use `-verbose` to see all errors.

**Q: Can I exclude specific files or directories?**
A: Yes, add them to `.gitignore` in the target directory. Alternatively, use `-ext` to include only specific extensions.
