## Context

`go-source-lookup` 是 Claude Code 中的 Go API 查詢工具。目前架構：
- 使用正則式從 LSP 警告/編譯錯誤中提取符號名
- 用 `go doc` 和 `go list` 查詢本地資訊
- 依賴 `$GOROOT` 和 `$GOPATH` 環境

問題：正則式解析精度低，無法處理複雜的 LSP 消息，且無法取得型別資訊。

## Goals / Non-Goals

**Goals:**
- 用 gopls LSP 替換現有的正則解析邏輯
- 支持基於檔案位置（file:line:col）的符號查詢
- 取得結構化的型別、簽名、文檔資訊
- 保留現有 CLI 接口的向後相容性
- 改進查詢精度和可靠性

**Non-Goals:**
- 修改 `query-remote.sh` 或遠程查詢邏輯
- 改變快取策略或 TTL
- 支持 LSP 以外的查詢方式
- 添加新的 CLI 命令（只擴展現有的）

## Decisions

### 決策 1：LSP 後端實現

**決定**：建立獨立的 `lsp-query.sh` 作為 LSP 查詢層，不改動現有的 `go-source-lookup.sh`。

**理由**：
- 保留現有邏輯作為降級方案
- 清晰的模塊責任分離
- 便於測試和偵錯

**替代方案**：直接改寫 `go-source-lookup.sh`（風險高，可能破壞現有功能）

### 決策 2：LSP 查詢方式

**決定**：使用 gopls 的 JSON-RPC 協議，透過 stdio 與 gopls 通訊。

**理由**：
- JSON-RPC 是標準 LSP 協議
- gopls 可靠且廣泛支持
- 無需額外依賴（Go 環境已包含）

**替代方案**：使用 VSCode LSP API（不適用 CLI 環境）

### 決策 3：快取策略

**決定**：基於 (file:line:col) 坐標快取 LSP 查詢結果，保留現有的 15 分鐘 TTL。

**理由**：
- 避免重複查詢同一位置
- 改善效能

### 決策 4：整合點

**決定**：在 `llm-integration.sh` 中新增 `--lsp-query` 模式，`trigger-logic.sh` 改為調用 LSP 查詢而非正則解析。

**理由**：
- 不破壞現有的 `--lsp` / `--compile-error` 接口
- 內部改用更精確的實現

## Risks / Trade-offs

| 風險 | 緩解方案 |
|------|---------|
| **gopls 不可用** | 檢查 `gopls --version`，降級到 `go doc` |
| **工作區配置錯誤** | 驗證 go.mod 存在並有效，提供清晰的錯誤消息 |
| **LSP 啟動延遲** | gopls 快取，重用單一進程實例 |
| **複雜的 LSP 消息格式** | 只提取必要的欄位（signature、documentation），忽略不相關欄位 |

## Migration Plan

1. **第 1 週**：實現 `lsp-query.sh` 和 LSP 通訊層
2. **第 2 週**：整合到 `trigger-logic.sh`，測試現有用例
3. **驗證**：對比舊實現和新實現的查詢結果精度
4. **上線**：無須服務停機，只是內部實現替換

## Open Questions

- 是否需要支持多個工作區（monorepo）？目前假設單工作區。
- LSP hover 返回的文檔格式是否需要特殊處理（Markdown vs 純文本）？
