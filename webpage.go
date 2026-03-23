package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/net/html"
)

type WebPageInput struct {
	URL string `json:"url" jsonschema:"The URL of the web page to download"`
}

func downloadWebPage(ctx context.Context, req *mcp.CallToolRequest, input WebPageInput) (
	*mcp.CallToolResult, any, error) {
	if input.URL == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "URL cannot be empty."},
			},
		}, nil, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(input.URL)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Failed to download web page."},
			},
		}, nil, nil
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Failed to read web page content."},
			},
		}, nil, nil
	}

	tokenizer := html.NewTokenizer(strings.NewReader(string(bodyBytes)))

	skip := false
	result := ""

	// A quick and WRONG way to extract text content from HTML by skipping certain tags.

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				break // End of document
			}
			log.Fatal(tokenizer.Err())
		}

		if tokenType == html.StartTagToken {
			token := tokenizer.Token()
			if token.Data == "script" || token.Data == "style" ||
				token.Data == "link" || token.Data == "meta" ||
				token.Data == "svg" || token.Data == "path" {
				skip = true
				continue
			}
		}

		if tokenType == html.TextToken {
			if skip {
				skip = false
				continue
			}

			token := tokenizer.Token()
			data := strings.TrimSpace(token.Data)
			if data != "" {
				result += html.UnescapeString(data) + " "
			}
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}
