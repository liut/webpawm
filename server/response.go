package server

// SearchResultResponse represents the response for search results
type SearchResultResponse struct {
	Engine  string                   `json:"engine"`
	Query   string                   `json:"query"`
	Count   int                      `json:"count"`
	Results []map[string]interface{} `json:"results"`
}
