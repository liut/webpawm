# 统一搜索工具设计

## Problem Frame

三个搜索工具（web_search、multi_search、smart_search）功能有重叠，用户体验碎片化。web_search 过于简单，multi_search 和 smart_search 各自有独特价值，但分开使用增加了复杂度。

目标是创建一个统一的搜索工具，同时满足：
- 简单场景：一句话搜索
- 多引擎场景：跨引擎比较/聚合
- 智能场景：查询扩展 + 去重 + 摘要

## Requirements

- R1. 单一工具名称为 `web_search`，替代现有的三个工具
- R2. **默认行为（智能模式）**：自动多引擎搜索 + 查询扩展 + 去重
- R3. 返回结构为**扁平列表 + summary**，简化调用方处理
- R4. **向后兼容**：现有调用方无需任何修改即可正常工作
- R5. 所有高级功能通过可选参数暴露，高级用户可按需调整

## Parameters

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `query` | string | **必填** | 搜索查询 |
| `engine` | string | `""` | 单一引擎（保留现有参数） |
| `engines` | []string | `[]` | 多引擎列表（新增，支持 multi_search 功能） |
| `max_results` | int | `10` | 最大结果数 |
| `language` | string | `""` | 语言筛选 |
| `arxiv_category` | string | `""` | 学术分类 |
| `search_depth` | string | `"normal"` | 搜索深度：quick/normal/deep（新增） |
| `include_academic` | bool | `false` | 包含学术搜索（新增） |
| `auto_query_expand` | bool | `true` | 自动查询扩展（新增） |
| `auto_deduplicate` | bool | `true` | 自动去重（新增） |

## Behavior

### 智能模式（默认）

当 `auto_query_expand=true`（默认）：
- `quick`: 1 个原始查询
- `normal`: +1 个新闻变体查询
- `deep`: +1 个学术变体查询

当 `auto_deduplicate=true`（默认）：
- 按 URL 去重后返回扁平列表

### 多引擎模式

当 `engines` 非空时：
- 忽略 `engine` 参数
- 并行在指定引擎执行查询
- 结果合并后去重（如果 `auto_deduplicate=true`）

### 简单模式

当 `auto_query_expand=false` 且 `engines` 为空时：
- 等同于原有 web_search 行为
- 单引擎 + 单次查询 + 不去重

## Response Structure

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
  "results": [
    {"index": 1, "title": "...", "link": "...", "snippet": "..."}
  ],
  "search_time": "2026-03-30T..."
}
```

## Scope Boundaries

- 移除 `multi_search` 工具
- 移除 `smart_search` 工具
- 不改变现有引擎接口（engine.Engine）
- 不改变 MCP 协议层

## Key Decisions

- **扁平结构**：始终返回去重后的扁平结果 + summary，避免调用方处理嵌套结构
- **智能默认**：智能模式作为默认行为，降低高级功能的使用门槛
- **参数正交**：`engine` 和 `engines` 互斥，`auto_query_expand` 控制查询扩展，各参数职责清晰

## Success Criteria

- [ ] 现有 web_search 调用方无需修改即可正常工作
- [ ] 新增 `engines` 参数可实现多引擎搜索
- [ ] 新增 `search_depth`/`auto_query_expand` 可实现查询扩展
- [ ] 新增 `auto_deduplicate` 可控制去重行为
- [ ] 返回结构始终为扁平列表 + summary
- [ ] multi_search 和 smart_search 工具可安全移除

## Dependencies / Assumptions

- 假设现有调用方主要使用 `query` 参数
- 假设 MCP 协议层支持可选参数的向后兼容

## Outstanding Questions

无

## Next Steps

→ `/ce:plan` for structured implementation planning
