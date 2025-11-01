# cc-token

A command-line tool for counting tokens in text files using Anthropic's Claude API.

## Description

`cc-token` is a simple CLI tool that reads a file and returns the token count for various Claude models. It uses Anthropic's token counting API to provide accurate token counts before sending content to Claude.

## Installation

### Using go install (Recommended)

```bash
go install github.com/iota-uz/cc-token@latest
```

This will install the `cc-token` binary to your `$GOPATH/bin` directory.

### From Source

```bash
git clone https://github.com/iota-uz/cc-token.git
cd cc-token
go build -o cc-token .
```

## Prerequisites

You need an Anthropic API key set as an environment variable:

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

## Usage

```bash
cc-token [flags] <path-to-file>
```

### Flags

- `-model` - Model to use for token counting (default: "claude-sonnet-4-5")

### Examples

Count tokens for a file using the default model:

```bash
cc-token document.txt
```

Count tokens using a specific model:

```bash
cc-token -model claude-opus-4 document.txt
```

Count tokens for a code file:

```bash
cc-token main.go
```

### Supported Models

- `claude-sonnet-4-5` (default)
- `claude-opus-4`
- `claude-haiku-4`
- Other Claude models supported by the Anthropic API

## Output

The tool outputs the file path and token count:

```
document.txt: 1234 tokens
```

## Error Handling

- If the file cannot be read, the tool will panic with an error message
- If the API request fails (invalid API key, network issues, etc.), the error response will be printed to stderr
- Ensure your `ANTHROPIC_API_KEY` environment variable is set correctly

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Links

- [Anthropic API Documentation](https://docs.anthropic.com/)
- [Token Counting API](https://docs.anthropic.com/en/docs/build-with-claude/token-counting)
