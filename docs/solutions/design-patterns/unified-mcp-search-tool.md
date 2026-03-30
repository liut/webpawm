---
title: "Unifying MCP Search Tools into a Single web_search Tool"
category: design-patterns
date: 2026-03-30
tags:
  - mcp
  - refactoring
  - api-design
  - search
summary: "Consolidated three fragmented search tools (web_search, multi_search, smart_search) into a single unified web_search tool with smart defaults. The new design introduces optional parameters for engines, search_depth, include_academic, auto_query_expand, and auto_deduplicate while maintaining backward-compatible intelligent behavior."
---

# Unifying MCP Search Tools into a Single web_search Tool

## Problem

Three search tools in the MCP server had overlapping functionality but different parameters and behaviors:
- **web_search**: Simple single-engine search
- **multi_search**: Multi-engine search without query expansion or deduplication
- **smart_search**: Intelligent query expansion but separate from the others

This fragmented design created cognitive burden for users and maintenance overhead.

## Root Cause

**Fragmented API design** — Instead of parameterizing similar capabilities, separate tools were created with different entry points.

## Solution

### 1. Unified Parameter Structure

Combined `WebSearchParams`, `MultiSearchParams`, and `SmartSearchParams` into a single struct:

```go
// server/params.go
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
```

### 2. Smart Defaults

| Parameter | Default | Behavior |
|-----------|---------|----------|
| `auto_query_expand` | `true` | Enable query expansion |
| `auto_deduplicate` | `true` | Deduplicate results by URL |
| `search_depth` | `"normal"` | 2 queries (original + news) |

### 3. Engine Resolution Logic

```go
// server/handlers.go:57-80
func (s *WebServer) resolveEngines(params WebSearchParams) []string {
    if len(params.Engines) > 0 {
        return params.Engines
    }
    if params.Engine != "" {
        return []string{params.Engine}
    }
    if params.AutoQueryExpand {
        engines := make([]string, 0, len(s.engines))
        for name := range s.engines {
            engines = append(engines, name)
        }
        return engines
    }
    return []string{s.defaultEngine}
}
```

### 4. Query Expansion by Depth

```go
// server/server.go:98-135
func generateSearchQueries(question, depth string) []searchQuery {
    queries := []searchQuery{}
    baseQueries := 1
    switch depth {
    case "quick":  baseQueries = 1
    case "normal": baseQueries = 2
    case "deep":   baseQueries = 3
    }
    queries = append(queries, searchQuery{Query: question, MaxResults: 10, Type: "general"})
    if baseQueries >= 2 {
        queries = append(queries, searchQuery{Query: question + " latest news", MaxResults: 5, Type: "news"})
    }
    if baseQueries >= 3 {
        queries = append(queries, searchQuery{Query: question + " research papers", MaxResults: 5, Type: "academic"})
    }
    return queries
}
```

### 5. Academic Query Filtering

```go
// server/handlers.go:146-152
func (s *WebServer) shouldUseEngine(engineName, queryType string, includeAcademic bool) bool {
    if queryType == "academic" {
        return includeAcademic && engineName == "arxiv"
    }
    return true
}
```

## Key Design Decisions

1. **Orthogonal parameters** — `engine` vs `engines` are mutually exclusive; `auto_query_expand` independently controls query expansion
2. **Backward compatibility** — Existing callers using only `query` get smart behavior by default
3. **Flat response structure** — Always returns `{summary, total_results, results, search_time}` regardless of mode

## Prevention

1. **Design for unification upfront** — When adding similar functionality, parameterize rather than create separate entry points
2. **Use orthogonal parameters** — Each parameter should control one aspect; avoid overlapping semantics
3. **Document default behaviors** — Ensure defaults are documented and backward-compatible
4. **Maintain decision rationale** — Keep plans in `docs/plans/` for future reference

## Related Documentation

- [Requirements doc](../brainstorms/2026-03-30-unified-search-requirements.md)
- [Implementation plan](../plans/2026-03-30-001-refactor-unified-search-plan.md)
