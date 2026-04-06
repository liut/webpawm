---
title: Analyze main.go to server package migration
type: refactor
status: active
date: 2026-04-06
---

# Analyze main.go to server package migration

## Overview

Analyze whether HTTP transport logic currently in `main.go` should be migrated to the `server/` package for better code organization and separation of concerns.

## Problem Statement

The `main.go` file (362 lines) contains a mix of concerns:
- CLI command definitions (cobra)
- Configuration loading (viper)
- **HTTP transport setup** (`startHTTPServer`, middleware)
- **Logger setup**
- Stdio server bootstrap

The `server/` package (163 lines in server.go alone, plus handlers.go, mcp.go, etc.) contains:
- WebServer struct and engine management
- MCP tool handlers
- Business logic

Currently, HTTP transport lives in `main.go` while the server business logic lives in `server/`. This creates a boundary where related code is split across packages.

## Current Architecture

```
main.go                           server/
┌────────────────────────────┐   ┌────────────────────────────┐
│ HTTP transport (lines 216-264)│   │ Business logic              │
│ - startHTTPServer()         │   │ - WebServer struct          │
│ - loggingMiddleware()       │   │ - Engine management         │
│ - apiKeyAuthMiddleware()    │   │ - MCP tool handlers         │
│ - setupLogger()              │   │ - Search orchestration      │
│ - responseWriter wrapper    │   │                             │
└────────────────────────────┘   └────────────────────────────┘
```

## Migration Candidates

### 1. HTTP Middleware (RECOMMENDED TO MIGRATE)

| Function | Location | Lines | Purpose |
|----------|----------|-------|---------|
| `loggingMiddleware` | main.go | 288-305 | HTTP request logging |
| `apiKeyAuthMiddleware` | main.go | 308-335 | API key authentication |
| `responseWriter` | main.go | 338-346 | Status code capture wrapper |

**Rationale:**
- Pure HTTP transport concerns
- No CLI-specific logic
- Self-contained and testable
- Related to `server/` which handles HTTP-based MCP

**Proposed new location:** `server/http.go` (new file)

### 2. Logger Setup (CONSIDER MIGRATING)

| Function | Location | Lines | Purpose |
|----------|----------|-------|---------|
| `setupLogger` | main.go | 267-285 | Creates slog.Logger from config |

**Rationale:**
- Used by HTTP middleware
- Could be part of server initialization
- However, stdio mode doesn't need HTTP logging

**Proposed new location:** `server/logger.go` (new file)

### 3. HTTP Server Bootstrap (DO NOT MIGRATE)

| Function | Location | Lines | Purpose |
|----------|----------|-------|---------|
| `startHTTPServer` | main.go | 216-264 | Full HTTP server setup |

**Rationale:**
- Orchestrates both `server/` (MCP) and `mcpserver` (HTTP transport)
- Contains CLI-specific print statements (`fmt.Printf`)
- References cobra command flags
- Lives at the intersection of CLI and server

**Decision: Keep in main.go** - This is the appropriate entry point for HTTP mode.

### 4. Stdio Server (DO NOT MIGRATE)

| Function | Location | Lines | Purpose |
|----------|----------|-------|---------|
| `runStdioServer` | main.go | 348-352 | Stdio mode bootstrap |
| `startStdioServer` | main.go | 354-362 | Stdio server startup |

**Decision: Keep in main.go** - Stdio is a transport like HTTP, but it makes sense to keep entry point logic in main.go.

## Proposed Migration

### Option A: Minimal Migration (Recommended)

Move only HTTP middleware to `server/`:

```
server/
├── http.go        # NEW: loggingMiddleware, apiKeyAuthMiddleware, responseWriter
├── logger.go      # NEW: setupLogger (optional)
├── server.go      # existing
├── handlers.go    # existing
├── mcp.go         # existing
└── ...
```

**Benefits:**
- Thin main.go (removes ~80 lines)
- HTTP middleware co-located with server HTTP handling
- Minimal changes, low risk

### Option B: Full Migration

Move middleware, logger setup, and create HTTP server factory in `server/`:

```
server/
├── http.go        # HTTP middleware + HTTP server factory
├── logger.go      # Logger setup
├── server.go      # existing + HTTP server creation
└── ...
```

**Benefits:**
- `main.go` becomes purely CLI bootstrap
- All server-related code in `server/`

**Drawbacks:**
- More complex refactoring
- `startHTTPServer` has CLI-specific output

## What NOT to Migrate

| Code | Reason to Keep |
|------|----------------|
| CLI commands (`rootCmd`, `webCmd`, etc.) | CLI entry point concerns |
| Viper config loading (`initConfig`, `getConfig`) | App bootstrap, not server |
| `startHTTPServer` orchestration | Mix of CLI output and server |
| `startStdioServer` | Entry point logic |

## Implementation Plan

### Phase 1: Create server/http.go

```go
// server/http.go
package server

import (
    "net/http"
    "time"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler { ... }

// APIKeyAuthMiddleware validates API keys
func APIKeyAuthMiddleware(validAPIKey string, next http.Handler, logger *slog.Logger) http.Handler { ... }

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct { ... }
```

### Phase 2: Create server/logger.go (optional)

```go
// server/logger.go
package server

import "log/slog"

// SetupLogger creates an slog.Logger based on LogLevel config
func SetupLogger(level string) *slog.Logger { ... }
```

### Phase 3: Update main.go

Remove migrated functions and update imports.

## Acceptance Criteria

- [ ] `server/http.go` contains all HTTP middleware
- [ ] `server/logger.go` contains logger setup (if migrated)
- [ ] `main.go` imports and uses migrated code from `server/`
- [ ] No functionality changes - refactor only
- [ ] Tests still pass
- [ ] HTTP and stdio modes both work correctly

## Files to Modify

| File | Action |
|------|--------|
| `server/http.go` | Create (new) |
| `server/logger.go` | Create (optional) |
| `main.go` | Update imports, remove migrated code |

## Files to Test

| File | What to Test |
|------|--------------|
| `server/http.go` | Middleware behavior unchanged |
| `main.go` | Both `web` and `std` commands work |

## Sources

- **Analysis target:** `main.go:216-346` (HTTP transport and middleware)
- **Analysis target:** `server/server.go:1-163` (server package structure)
- **Repo research:** Technology stack Go 1.24.0, cobra v1.10.2, viper v1.21.0, mcp-go v0.46.0
