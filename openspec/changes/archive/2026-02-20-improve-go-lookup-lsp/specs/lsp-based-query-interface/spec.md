## ADDED Requirements

### Requirement: 基於檔案位置的 LSP 查詢接口

系統應提供新的查詢接口，支持基於具體檔案位置（檔案路徑、行號、列號）進行 LSP 符號查詢。

#### Scenario: 使用檔案位置查詢符號

- **WHEN** 用戶調用 `llm-integration.sh --lsp-query <file> <line> <col>`
- **THEN** 系統連接到 gopls，發送 `textDocument/hover` 請求
- **THEN** 系統返回該位置的符號詳細資訊

#### Scenario: 處理無效檔案位置

- **WHEN** 提供的檔案不存在或行號超出範圍
- **THEN** 系統返回清晰的錯誤消息，包括檔案路徑和位置

#### Scenario: 查詢結果格式化

- **WHEN** LSP 查詢成功且返回懸停信息
- **THEN** 系統格式化為 Markdown，包含簽名、文檔、型別資訊

### Requirement: 支持多種 LSP 查詢方法

系統應支持多種 LSP 查詢方法以應對不同場景。

#### Scenario: 查詢定義位置

- **WHEN** 用戶需要找到符號的定義位置
- **THEN** 系統使用 `textDocument/definition` 返回檔案位置和範圍

#### Scenario: 查詢符號引用

- **WHEN** 用戶需要找出符號在代碼中被引用的位置
- **THEN** 系統使用 `textDocument/references` 返回所有引用位置清單

#### Scenario: 查詢型別資訊

- **WHEN** 用戶懸停在變量或表達式上
- **THEN** 系統透過 LSP hover 返回推導的型別資訊

### Requirement: 向後相容現有 CLI 接口

系統應保持現有 CLI 接口的相容性，內部改用 LSP 查詢。

#### Scenario: 舊接口繼續工作

- **WHEN** 用戶使用 `go-source-lookup.sh <package> [symbol]`
- **THEN** 系統仍能正常工作，內部改用 LSP 或 `go doc` 混合查詢

#### Scenario: 自動觸發查詢保持功能

- **WHEN** LSP 警告或編譯錯誤被觸發
- **THEN** 系統自動調用 LSP 查詢（而非正則解析）並返回結果

#### Scenario: 快取鍵匹配

- **WHEN** 新舊查詢方式查詢同一符號
- **THEN** 系統能正確識別快取重複，避免不必要的查詢
