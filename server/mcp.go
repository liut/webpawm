package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CreateMcpServer creates the MCP server with tools
func (s *WebServer) CreateMcpServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "webpawm",
		Version: "1.0.0",
	}, nil)

	// Add web_search tool (unified: supports single engine, multi-engine, and smart search)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "web_search",
		Description: "Search the web using various search engines. Supports single engine, multi-engine parallel search, and intelligent query expansion with deduplication by default.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "The search query",
				},
				"engine": map[string]any{
					"type":        "string",
					"description": "Single search engine to use (mutually exclusive with engines)",
					"enum":        s.getAvailableEngines(),
				},
				"engines": map[string]any{
					"type":        "array",
					"description": "List of search engines to use (mutually exclusive with engine). Uses all available engines if empty with auto_query_expand enabled",
					"items": map[string]any{
						"type": "string",
						"enum": s.getAvailableEngines(),
					},
				},
				"max_results": map[string]any{
					"type":        "integer",
					"description": "Maximum number of results to return (default: 10)",
					"minimum":     1,
					"maximum":     50,
				},
				"language": map[string]any{
					"type":        "string",
					"description": "Language code for search results (e.g., 'en', 'zh')",
				},
				"arxiv_category": map[string]any{
					"type":        "string",
					"description": "Arxiv category for academic paper search (e.g., 'cs.AI', 'math.CO')",
				},
				"search_depth": map[string]any{
					"type":        "string",
					"description": "Search depth: 'quick' (1 query), 'normal' (2 queries), 'deep' (3 queries). Default: 'normal'",
					"enum":        []string{"quick", "normal", "deep"},
				},
				"include_academic": map[string]any{
					"type":        "boolean",
					"description": "Include academic papers from Arxiv (default: false)",
				},
				"auto_query_expand": map[string]any{
					"type":        "boolean",
					"description": "Automatically expand query with variations (news, academic) based on search_depth (default: true)",
				},
				"auto_deduplicate": map[string]any{
					"type":        "boolean",
					"description": "Automatically deduplicate results by URL (default: true)",
				},
			},
			"required": []string{"query"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, params WebSearchParams) (*mcp.CallToolResult, any, error) {
		result, err := s.handleWebSearch(ctx, params)
		return result, nil, err
	})

	// Add web_fetch tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "web_fetch",
		Description: "Fetch a website and return its content. Supports HTML to Markdown conversion for readability.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url": map[string]any{
					"type":        "string",
					"description": "URL of the website to fetch",
				},
				"max_length": map[string]any{
					"type":        "integer",
					"description": "Maximum number of characters to return (default: 5000)",
					"minimum":     1,
					"maximum":     999999,
				},
				"start_index": map[string]any{
					"type":        "integer",
					"description": "Start content from this character index (default: 0)",
					"minimum":     0,
				},
				"raw": map[string]any{
					"type":        "boolean",
					"description": "If true, returns the raw HTML including <script> and <style> blocks (default: false)",
				},
			},
			"required": []string{"url"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, params WebFetchParams) (*mcp.CallToolResult, any, error) {
		result, err := s.handleWebFetch(ctx, params)
		return result, nil, err
	})

	return server
}

// Run starts the MCP server over SSE
func (s *WebServer) Run(addr string) error {
	mcpServer := s.CreateMcpServer()

	handler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
		return mcpServer
	}, nil)

	fmt.Printf("Webpawm MCP server starting on %s\n", addr)
	return http.ListenAndServe(addr, handler)
}
