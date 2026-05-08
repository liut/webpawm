---
name: webpawm-cli
description: Web search and web fetch via CLI. Use when searching the web or fetching page content from the command line.
---

# Webpawm CLI - Web Search & Fetch (Command Line)

Provides two CLI subcommands: **search** (multi-engine web search) and **fetch** (HTML→Markdown content extraction).

## Engine Availability

| Engine | Always available? | Requires config |
|--------|-------------------|-----------------|
| `bingcn` | Yes | None |
| `arxiv` | Yes | None |
| `google` | No | `google_api_key` + `google_cx` |
| `bing` | No | `bing_api_key` |
| `brave` | No | `brave_api_key` |
| `searchxng` | No | `searchxng_url` |

Use `bingcn` or `arxiv` for instant testing. Other engines require API keys in `~/.config/webpawm/config.json`.

## Expected Latency

| Search depth | Engines | Typical time |
|-------------|---------|--------------|
| `quick` | 1 engine, 1 query | 1–5s |
| `normal` | 1 engine, 2 queries | 3–10s |
| `deep` | 1 engine, 3 queries | 5–20s |
| Multi-engine | 2–3 engines | 5–30s |

Fetch typically takes 1–10s depending on target page size and network. HTTP client timeout is 30s.

---

## Search

```bash
webpawm search "your search query" [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `-e, --engine` | Single search engine | config default |
| `--engines` | Multiple engines (comma-separated) | — |
| `-n, --max-results` | Max results to return | config value |
| `-l, --language` | Language code (`en`, `zh`, etc.) | — |
| `--arxiv-category` | Arxiv category (`cs.AI`, `math.CO`, etc.) | — |
| `-d, --depth` | Search depth: `quick`/`normal`/`deep` | `normal` |
| `--academic` | Include academic papers from Arxiv | `false` |
| `--no-expand` | **Disable** auto query expansion (expansion is ON by default) | `false` |
| `--no-dedup` | **Disable** auto result deduplication (dedup is ON by default) | `false` |

> Boolean semantics: `auto_query_expand` and `auto_deduplicate` default to `true`. The CLI uses inverted flags: `--no-expand` disables expansion, `--no-dedup` disables deduplication. Not passing these flags means the feature is ON.

## Fetch

```bash
webpawm fetch "https://example.com" [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `-n, --max-length` | Max characters to return | `5000` |
| `-s, --start-index` | Start from this character index | `0` |
| `-r, --raw` | Return raw HTML instead of Markdown | `false` |

---

## Success Response: Search

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
      "title": "An Introduction To Generics - The Go Programming Language",
      "link": "https://go.dev/doc/tutorial/generics",
      "snippet": "This tutorial introduces the basics of generics in Go..."
    }
  ],
  "search_time": "2026-05-07T00:00:00+08:00"
}
```

## Success Response: Fetch

```json
{
  "url": "https://example.com",
  "content": "# Example Domain\n\nThis domain is for use in illustrative examples...",
  "content_type": "markdown",
  "original_length": 1250,
  "truncated": false,
  "next_start": 0
}
```

When `truncated` is `true`, `next_start` contains the next `start_index` to pass for pagination.

When fetch fails (HTTP error, network issue), error details are in the `error` field:
```json
{"url": "https://broken.example.com", "content": "", "content_type": "", "original_length": 0, "truncated": false, "error": "Error fetching URL: HTTP 404"}
```

---

## Examples

```bash
# Basic search
webpawm search "golang generics tutorial"

# Search with specific engine
webpawm search -e google "climate change"

# Deep search with academic papers
webpawm search -d deep --academic "transformer architecture"

# Multi-engine search
webpawm search --engines google,bing "webassembly"

# Fetch a web page as Markdown
webpawm fetch "https://go.dev/doc/tutorial/getting-started"

# Fetch raw HTML
webpawm fetch -r "https://example.com"

# Fetch with pagination
webpawm fetch -n 5000 -s 5000 "https://example.com/large-page"
```
