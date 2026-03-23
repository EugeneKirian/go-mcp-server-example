package main

import (
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-server",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_time",
		Description: "Get the current time in a specified timezone",
	}, getTime)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_search_results",
		Description: "Get internet search results for a query",
	}, getSearchResults)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "download_web_page",
		Description: "Download and extract text content of a web page given its URL",
	}, downloadWebPage)

	// Run server on http transport
	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{},
	)

	http.HandleFunc("/mcp", handler.ServeHTTP)

	log.Println("Starting MCP server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
