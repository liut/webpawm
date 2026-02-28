package server

import (
	"slices"

	"github.com/liut/wisper/engine"
)

// Config for the web search server.
// Can be loaded from environment variables using ProcessConfig.
// Supports configuration from file (~/.wisper/config.json) and environment variables.
type Config struct {
	SearchXNGURL  string `mapstructure:"searchxng_url" envconfig:"SEARCHXNG_URL"`    // SearXNG base URL (e.g., https://searchx.ng)
	GoogleAPIKey  string `mapstructure:"google_api_key" envconfig:"GOOGLE_API_KEY"`  // Google Custom Search API key
	GoogleCX      string `mapstructure:"google_cx" envconfig:"GOOGLE_CX"`            // Google Search Engine ID
	BingAPIKey    string `mapstructure:"bing_api_key" envconfig:"BING_API_KEY"`      // Bing Search API key
	MaxResults    int    `mapstructure:"max_results" envconfig:"MAX_RESULTS"`         // Default max results (default: 10)
	DefaultEngine string `mapstructure:"default_engine" envconfig:"DEFAULT_ENGINE"`    // Default search engine
	ListenAddr    string `mapstructure:"listen_addr" envconfig:"LISTEN_ADDR"`         // HTTP listen address
	URIPrefix     string `mapstructure:"uri_prefix" envconfig:"URI_PREFIX"`           // URI prefix for HTTP endpoints
}

// WebSearchServer represents the MCP web search server
type WebSearchServer struct {
	engines       map[string]engine.Engine
	defaultEngine string
	maxResults    int
}

// NewWebSearchServer creates a new web search server
func NewWebSearchServer(config Config) *WebSearchServer {
	engines := make(map[string]engine.Engine)

	// Add Google engine if configured
	if config.GoogleAPIKey != "" && config.GoogleCX != "" {
		engines["google"] = engine.NewGoogleEngine(config.GoogleAPIKey, config.GoogleCX)
	}

	// Add Bing engine if configured
	if config.BingAPIKey != "" {
		engines["bing"] = engine.NewBingEngine(config.BingAPIKey)
	}

	// Add Bing CN engine (free, always available)
	engines["bingcn"] = engine.NewBingCNEngine()

	// Add SearXNG engine if configured
	if config.SearchXNGURL != "" {
		engines["searchxng"] = engine.NewSearchXNGEngine(config.SearchXNGURL)
	}

	// Add Arxiv engine (always available)
	engines["arxiv"] = engine.NewArxivEngine()

	defaultEngine := config.DefaultEngine
	if defaultEngine == "" && len(engines) > 0 {
		for name := range engines {
			defaultEngine = name
			break
		}
	}

	maxResults := config.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}

	return &WebSearchServer{
		engines:       engines,
		defaultEngine: defaultEngine,
		maxResults:    maxResults,
	}
}

// MustNewWebSearchServer creates a new web search server with error handling
func MustNewWebSearchServer(searchXNGURL string, maxResults int) *WebSearchServer {
	config := Config{
		SearchXNGURL: searchXNGURL,
		MaxResults:   maxResults,
	}
	return NewWebSearchServer(config)
}

// searchQuery represents a search query with parameters
type searchQuery struct {
	Query      string
	MaxResults int
	Type       string // "general", "academic", "news"
}

// generateSearchQueries creates search queries based on the user's question and depth
func generateSearchQueries(question, depth string) []searchQuery {
	queries := []searchQuery{}

	baseQueries := 1
	switch depth {
	case "quick":
		baseQueries = 1
	case "normal":
		baseQueries = 2
	case "deep":
		baseQueries = 3
	}

	queries = append(queries, searchQuery{
		Query:      question,
		MaxResults: 10,
		Type:       "general",
	})

	if baseQueries >= 2 {
		queries = append(queries, searchQuery{
			Query:      question + " latest news",
			MaxResults: 5,
			Type:       "news",
		})
	}

	if baseQueries >= 3 {
		queries = append(queries, searchQuery{
			Query:      question + " research papers",
			MaxResults: 5,
			Type:       "academic",
		})
	}

	return queries
}

// determineEngine selects the appropriate search engine for a query
func (s *WebSearchServer) determineEngine(queryType string, includeAcademic bool) string {
	if queryType == "academic" && includeAcademic {
		if _, ok := s.engines["arxiv"]; ok {
			return "arxiv"
		}
	}

	if _, ok := s.engines["searchxng"]; ok {
		return "searchxng"
	}

	for name := range s.engines {
		return name
	}

	return ""
}

// removeDuplicates removes duplicate search results based on URL
func (s *WebSearchServer) removeDuplicates(results []engine.SearchResult) []engine.SearchResult {
	seen := make(map[string]bool, len(results))
	var unique []engine.SearchResult

	for _, result := range results {
		if !seen[result.Link] {
			seen[result.Link] = true
			unique = append(unique, result)
		}
	}

	return unique
}

// getAvailableEngines returns a list of available search engine names
func (s *WebSearchServer) getAvailableEngines() []string {
	names := make([]string, 0, len(s.engines))
	for name := range s.engines {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
