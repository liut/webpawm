---
title: refactor: Replace webFetchOutput with typed WebFetchResponse struct
type: refactor
status: completed
date: 2026-04-06
---

# refactor: Replace webFetchOutput with typed WebFetchResponse struct

## Problem Statement

`handleWebFetch` returns `*webFetchOutput` which only contains a `Text` string. This lacks type safety and forces clients to parse a formatted string to extract information like URL, content type, and truncation status.

## Current vs Proposed Structure

| Current | Proposed |
|---------|----------|
| `webFetchOutput{Text: "Markdown\nContents of URL:\ncontent..."}` | `WebFetchResponse{URL, Content, ContentType, OriginalLength, Truncated, NextStart, Error}` |

## Current Call Chain

```
handleWebFetchHandler (mcp.go)
  └── returns webFetchOutput {Text: string}
      └── calls handleWebFetch (handle_fetch.go)
              └── returns *webFetchOutput {Text: formatted string}
```

## Proposed Call Chain

```
handleWebFetchHandler (mcp.go)
  └── returns *WebFetchResponse ✅
      └── calls handleWebFetch
              └── returns *WebFetchResponse ✅
```

## Changes Summary

| File | Change |
|------|--------|
| `server/response.go` | Add `WebFetchResponse` struct |
| `server/handle_fetch.go` | Update `handleWebFetch` to return `*WebFetchResponse` |
| `server/mcp.go` | Update `handleWebFetchHandler` return type |
| `server/params.go` | Remove `webFetchOutput` |

### New Struct (in `server/response.go`)

```go
type WebFetchResponse struct {
    URL            string `json:"url"`
    Content        string `json:"content"`
    ContentType    string `json:"content_type"` // "markdown" or "raw"
    OriginalLength int    `json:"original_length"`
    Truncated      bool   `json:"truncated"`
    NextStart      int    `json:"next_start,omitempty"`
    Error          string `json:"error,omitempty"`
}
```

## Acceptance Criteria

- [ ] `WebFetchResponse` struct defined in `server/response.go`
- [ ] `handleWebFetch` returns `*WebFetchResponse`
- [ ] `handleWebFetchHandler` returns `*WebFetchResponse`
- [ ] `webFetchOutput` removed from `params.go`
- [ ] Code compiles
- [ ] MCP server functions correctly (build verified)
- [ ] Existing tests pass

## Benefits

- **Type safety** - Compile-time checking throughout the call chain
- **Structured data** - Clients can directly access URL, content type, truncation info
- **Self-documenting** - JSON field names document the response structure
- **Better error handling** - Explicit error field instead of parsing error strings

## Files

- `server/response.go` - Add `WebFetchResponse` struct
- `server/handle_fetch.go` - Update `handleWebFetch` return type
- `server/mcp.go` - Update `handleWebFetchHandler` return type
- `server/params.go` - Remove `webFetchOutput`
