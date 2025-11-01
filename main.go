package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	model := flag.String("model", "claude-sonnet-4-5", "Model to use for token counting")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: cc-token [flags] <path-to-file>")
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	fp := flag.Arg(0)
	b, err := os.ReadFile(fp)
	if err != nil {
		panic(err)
	}

	body := map[string]any{
		"model": *model,
		"messages": []map[string]any{
			{"role": "user", "content": string(b)},
		},
	}
	j, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages/count_tokens", bytes.NewReader(j))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("x-api-key", os.Getenv("ANTHROPIC_API_KEY"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		io.Copy(os.Stderr, resp.Body)
		os.Exit(1)
	}

	var out struct {
		InputTokens int `json:"input_tokens"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		panic(err)
	}
	fmt.Printf("%s: %d tokens\n", fp, out.InputTokens)
}
