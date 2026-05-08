---
name: webpawm-api
description: Web search and web fetch via REST API. Use when interacting with a running webpawm server over HTTP.
---

# Webpawm API - Web Search & Fetch (REST API)

Provides REST API endpoints for **web search** (multi-engine) and **web fetch** (HTML→Markdown) when a webpawm server is running.

Start the server: `webpawm web -l localhost:8087`

## Engine Availability

| Engine | Always available? | Requires config |
|--------|-------------------|-----------------|
| `bingcn` | Yes | None |
| `arxiv` | Yes | None |
| `google` | No | `google_api_key` + `google_cx` |
| `bing` | No | `bing_api_key` |
| `brave` | No | `brave_api_key` |
| `searchxng` | No | `searchxng_url` |

---

## Health & Discovery

```bash
# Server health check
curl -s http://localhost:8087/api/health
# → {"status":"ok","version":"1.0.0"}

# List configured engines
curl -s http://localhost:8087/api/engines
# → {"engines":["arxiv","bingcn","google"],"default":"bingcn"}
```

## Search

```bash
curl -s -X POST http://localhost:8087/api/search \
  -H "Content-Type: application/json" \
  -d '{"query": "golang generics", "engine": "google", "max_results": 5}'
```

### Search Request Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `query` | string | yes | Search query |
| `engine` | string | no | Single engine name |
| `engines` | []string | no | Multiple engine names |
| `max_results` | int | no | Max results (default: 10) |
| `language` | string | no | Language code |
| `arxiv_category` | string | no | Arxiv category |
| `search_depth` | string | no | `quick`/`normal`/`deep` |
| `include_academic` | bool | no | Include academic papers (default: false) |
| `auto_query_expand` | bool | no | Auto expand queries (default: true) |
| `auto_deduplicate` | bool | no | Auto deduplicate results (default: true) |

### Search Success Response

```json
{
  "total_results": 3,
  "summary": {
    "total_raw_results": 3,
    "total_unique_results": 3,
    "original_query": "golang generics",
    "search_queries": ["golang generics", "golang generics latest news"],
    "engines_used": ["bingcn"],
    "search_depth": "normal"
  },
  "results": [
    {
      "index": 1,
      "title": "Result Title",
      "link": "https://example.com/result",
      "snippet": "Result snippet text..."
    }
  ],
  "search_time": "2026-05-07T00:00:00+08:00"
}
```

## Fetch

```bash
curl -s -X POST http://localhost:8087/api/fetch \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "max_length": 3000}'
```

### Fetch Request Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `url` | string | yes | Website URL (http/https only) |
| `max_length` | int | no | Max chars (default: 5000) |
| `start_index` | int | no | Start char index (default: 0) |
| `raw` | bool | no | Return raw HTML (default: false) |

### Fetch Success Response

```json
{
  "url": "https://example.com",
  "content": "# Markdown content of the page...",
  "content_type": "markdown",
  "original_length": 1250,
  "truncated": false,
  "next_start": 0,
  "error": ""
}
```

**Pagination:** When `truncated` is `true`, pass `next_start` as `start_index` in the next request to continue reading.

**Fetch errors:** On HTTP errors (404, timeout, etc.), the response has `"error"` populated with details. The HTTP status is still 200. Check the `error` field, not the HTTP status code.

---

## Authentication

If `api_key` is configured, include it:

```bash
curl -s -X POST http://localhost:8087/api/search \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"query": "test"}'
```

---

## Error Responses

```json
{"error": "Bad Request", "message": "query is required"}
{"error": "Unauthorized", "message": "API key required..."}
{"error": "Payload Too Large", "message": "request body exceeds 1MB limit"}
{"error": "Unsupported Media Type", "message": "Content-Type must be application/json"}
```

HTTP status codes:
- `200` — success (including fetch HTTP errors, check `error` field)
- `400` — validation error
- `401` — auth required
- `413` — body too large
- `415` — not JSON
- `500` — server error
