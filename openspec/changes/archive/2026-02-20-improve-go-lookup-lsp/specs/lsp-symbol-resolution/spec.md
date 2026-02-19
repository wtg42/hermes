## ADDED Requirements

### Requirement: 通過 LSP 查詢本地符號

系統應能使用 gopls 的 LSP 協議查詢本地 Go 符號，提供準確的型別、簽名和文檔資訊。

#### Scenario: 查詢標準庫符號

- **WHEN** 用戶查詢標準庫符號（例如 `fmt.Println`）
- **THEN** 系統透過 gopls 返回完整的函數簽名、參數型別和文檔

#### Scenario: 查詢依賴包符號

- **WHEN** 用戶查詢已安裝依賴中的符號
- **THEN** 系統透過 LSP 返回符號的準確型別和來源位置

#### Scenario: LSP 服務不可用時降級

- **WHEN** gopls 不可用或啟動失敗
- **THEN** 系統降級回 `go doc` 查詢，提供有限的文檔資訊

### Requirement: 提取並格式化 LSP 懸停資訊

系統應從 LSP hover 回應中提取關鍵資訊，格式化為 LLM 易讀的形式。

#### Scenario: 提取函數簽名

- **WHEN** LSP hover 返回函數定義
- **THEN** 系統提取並正規化簽名格式（例如 `func (r *Reader) Read(b []byte) (n int, err error)`)

#### Scenario: 提取文檔字符串

- **WHEN** LSP hover 返回包含文檔的回應
- **THEN** 系統提取純文本文檔（去除 Markdown 語法）並格式化

#### Scenario: 處理已廢棄 API

- **WHEN** LSP hover 檢測到已廢棄的符號（含 Deprecated 標記）
- **THEN** 系統在結果中標記為「已廢棄」並建議替代方案

### Requirement: 緩存 LSP 查詢結果

系統應基於查詢坐標（檔案:行:列）緩存 LSP 查詢結果，減少重複查詢。

#### Scenario: 命中緩存結果

- **WHEN** 用戶在同一位置多次查詢
- **THEN** 系統從緩存返回結果，無需重新調用 LSP

#### Scenario: 緩存過期時更新

- **WHEN** 快取超過 15 分鐘
- **THEN** 系統丟棄舊緩存，重新查詢 LSP
