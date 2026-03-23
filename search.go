package main

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SearchInput struct {
	Query string `json:"query" jsonschema:"Search query"`
}

type SearchResult struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func getSearchResults(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (
	*mcp.CallToolResult, any, error) {
	if input.Query == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Query cannot be empty."},
			},
		}, nil, nil
	}

	search := NewSearcher(input.Query)
	results, err := search.Search()

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error performing search: " + err.Error()},
			},
		}, nil, nil
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Cannot marshal search results."},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonData)},
		},
	}, nil, nil
}
