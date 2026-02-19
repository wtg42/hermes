## Why

目前 `go-source-lookup` 在本地查詢時依賴正則表達式解析 LSP 警告和編譯錯誤，導致精度不足。通過改用 gopls 的 LSP 協議，可以取得結構化、準確的符號資訊、型別和簽名，提升查詢品質。

## What Changes

- **替換符號解析邏輯**：從正則式改用 `gopls` LSP 的 `textDocument/hover` 和 `textDocument/definition`
- **新增型別推導**：通過 LSP hover 取得完整型別資訊
- **改進簽名取得**：直接從 LSP 取得準確的函數簽名，而非依賴 `go doc` 的字符串輸出
- **保留快取機制**：維持現有的 15 分鐘快取策略
- **保留遠程查詢**：官方文件和遠程包查詢保持現狀

## Capabilities

### New Capabilities

- `lsp-symbol-resolution`：使用 gopls LSP 協議查詢本地符號，取得準確的型別、簽名和文檔
- `lsp-based-query-interface`：新 API 接口支持基於檔案位置的符號查詢（檔案路徑、行號、列號）

### Modified Capabilities

- `go-source-lookup`：現有的本地查詢邏輯將升級為使用 LSP 後端

## Impact

- **文件**：`go-source-lookup.sh`、`trigger-logic.sh`、`llm-integration.sh`
- **新依賴**：需要 gopls（通常已隨 Go 環境提供）
- **API 變化**：`llm-integration.sh` 新增 `--lsp-query <file> <line> <col>` 選項
- **向後相容**：現有的 CLI 接口保持相同，內部改用 LSP 後端
