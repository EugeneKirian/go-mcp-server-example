package main

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TimeInput struct {
	Timezone string `json:"timezone" jsonschema:"IANA timezone identifier (e.g. America/New_York)"`
}

func getTime(ctx context.Context, req *mcp.CallToolRequest, input TimeInput) (
	*mcp.CallToolResult, any, error) {
	loc, err := time.LoadLocation(input.Timezone)

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Cannot load Timezone: " + err.Error()},
			},
		}, nil, nil
	}

	timeStamp := time.Now().In(loc)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Time is:" + timeStamp.Format(time.RFC3339)},
		},
	}, nil, nil
}
