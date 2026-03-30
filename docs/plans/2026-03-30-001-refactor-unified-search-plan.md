---
title: refactor: Unified search tool (web_search + multi_search + smart_search)
type: refactor
status: active
date: 2026-03-30
origin: docs/brainstorms/2026-03-30-unified-search-requirements.md
---

# Unified Search Tool

## Overview

将三个搜索工具（web_search、multi_search、smart_search）合并为单一的 `web_search` 工具，默认启用智能模式（多引擎 + 查询扩展 + 去重），同时通过可选参数暴露高级控制能力。

## Problem Statement

现有三个工具功能重叠但体验碎片化：
- `web_search`：简单但功能有限
- `multi_search`：多引擎搜索能力，但无查询扩展/去重
- `smart_search`：智能查询扩展，但与其他两个工具分离

用户需要记忆三个工具的不同参数和行为，增加认知负担。

## Proposed Solution

改造 `web_search` 为统一工具：
1. **默认智能模式**：多引擎 + 查询扩展 + 去重
2. **可选参数暴露高级功能**：engines, search_depth, auto_query_expand, auto_deduplicate
3. **始终返回扁平结构 + summary**

## Technical Approach

### 修改范围

| 文件 | 修改内容 |
|------|----------|
| `server/params.go` | 更新 `WebSearchParams` 结构体，新增参数 |
| `server/mcp.go` | 更新 `web_search` 工具的 InputSchema；移除 `multi_search` 和 `smart_search` 工具注册 |
| `server/handlers.go` | 重写 `handleWebSearch` 实现统一逻辑 |
| `server/server.go` | 可能需要调整 `generateSearchQueries`、`determineEngine`、`removeDuplicates` 的可见性或逻辑 |

### 实现步骤

#### Step 1: 更新参数结构体 (`server/params.go`)

新增以下字段到 `WebSearchParams`：

```go
type WebSearchParams struct {
    Query         string   `json:"query"`
    Engine        string   `json:"engine"`           // 保留
    Engines       []string `json:"engines"`          // 新增
    MaxResults    int      `json:"max_results"`
    Language      string   `json:"language"`
    ArxivCategory string   `json:"arxiv_category"`
    SearchDepth   string   `json:"search_depth"`     // 新增: quick/normal/deep
    IncludeAcademic bool   `json:"include_academic"` // 新增
    AutoQueryExpand bool   `json:"auto_query_expand"` // 新增: 默认 true
    AutoDeduplicate bool   `json:"auto_deduplicate"`  // 新增: 默认 true
}
```

#### Step 2: 更新工具注册 (`server/mcp.go`)

**web_search 工具**：更新 InputSchema 以包含新参数

**移除工具注册**：
- 删除 `multi_search` 工具注册（ lines 57-87）
- 删除 `smart_search` 工具注册（lines 89-115）

#### Step 3: 重写 handleWebSearch (`server/handlers.go`)

实现统一逻辑：

```go
func (s *WebServer) handleWebSearch(ctx context.Context, params WebSearchParams) (*mcp.CallToolResult, error) {
    // 1. 参数校验
    if params.Query == "" {
        return nil, errors.New("query is required")
    }

    // 2. 确定使用的引擎列表
    engines := s.resolveEngines(params)  // 新增 helper

    // 3. 确定是否启用查询扩展
    queries := []string{params.Query}
    if params.AutoQueryExpand {
        queries = generateSearchQueries(params.Query, params.SearchDepth)
    }

    // 4. 执行搜索
    allResults := s.executeSearches(ctx, engines, queries, params)

    // 5. 去重（如启用）
    if params.AutoDeduplicate {
        allResults = s.removeDuplicates(allResults)
    }

    // 6. 格式化响应
    return s.formatResponse(params, engines, queries, allResults)
}
```

关键 helper 函数：
- `resolveEngines(params)`: 根据 `engines` 和 `engine` 参数确定引擎列表
- `executeSearches(ctx, engines, queries, params)`: 并行执行搜索
- `formatResponse(params, engines, queries, results)`: 生成扁平响应结构

### 向后兼容保证

- `query` 仍为必填参数
- `engine` 参数语义不变
- 默认值确保现有调用方行为不变：
  - `AutoQueryExpand=true`（智能扩展默认启用）
  - `AutoDeduplicate=true`（去重默认启用）
  - `search_depth="normal"`（中等深度）

### 响应结构

始终返回（见 origin doc）：
```json
{
  "summary": {
    "original_query": "...",
    "search_queries": ["..."],
    "engines_used": ["..."],
    "search_depth": "normal",
    "total_raw_results": 25,
    "total_unique_results": 18
  },
  "total_results": 18,
  "results": [...],
  "search_time": "..."
}
```

## System-Wide Impact

- **移除的工具**：`multi_search`、`smart_search` 不再可用
- **MCP 协议**：web_search InputSchema 变更，但向后兼容（新增可选参数）
- **引擎接口**：无需修改

## Acceptance Criteria

- [ ] R1: 单一工具名称为 `web_search`
- [ ] R2: 默认行为（智能模式）正常工作
- [ ] R3: 返回扁平结构 + summary
- [ ] R4: 向后兼容（现有调用方无需修改）
- [ ] R5: 所有高级功能通过可选参数可调
- [ ] 移除 `multi_search` 工具
- [ ] 移除 `smart_search` 工具

## Dependencies & Risks

- 依赖 MCP SDK v1.4.0（已确认支持可选参数）
- 风险：多引擎并行搜索可能增加延迟（但结果是聚合的，更全面）

## Sources

- **Origin document:** docs/brainstorms/2026-03-30-unified-search-requirements.md
  - Key decisions: 智能默认、扁平结构、参数正交
- 现有模式复用：
  - `server/server.go:98-169` - generateSearchQueries, determineEngine, removeDuplicates
  - `server/handlers.go:78-143` - handleMultiSearch 多引擎聚合逻辑
  - `server/handlers.go:146-216` - handleSmartSearch 查询扩展逻辑
