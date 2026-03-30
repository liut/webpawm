package server

// WebSearchParams represents the parameters for web_search tool
type WebSearchParams struct {
	Query           string   `json:"query"`
	Engine          string   `json:"engine"`
	Engines         []string `json:"engines"`
	MaxResults      int      `json:"max_results"`
	Language        string   `json:"language"`
	ArxivCategory   string   `json:"arxiv_category"`
	SearchDepth     string   `json:"search_depth"`
	IncludeAcademic bool     `json:"include_academic"`
	AutoQueryExpand bool     `json:"auto_query_expand"`
	AutoDeduplicate bool     `json:"auto_deduplicate"`
}

// WebFetchParams represents the parameters for web_fetch tool
type WebFetchParams struct {
	URL       string `json:"url"`
	MaxLength int    `json:"max_length"`
	StartIndex int   `json:"start_index"`
	Raw       bool   `json:"raw"`
}
