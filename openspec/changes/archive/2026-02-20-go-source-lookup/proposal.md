## Why

當 Claude Code 輔助編寫 Go 代碼時，會遇到 LSP 報警、編譯錯誤、或使用廢棄 API 等問題。此時 LLM 需要快速查詢最新的標準庫和依賴套件實現，而不是依賴訓練數據中可能過時的 API 知識。這個 skill 讓 LLM 能夠自動檢索本機 Go 源碼和遠端套件，確保生成的代碼符合當前環境和最新最佳實踐。

## What Changes

- 新增 `go-source-lookup` skill，賦予 Claude Code 查詢 Go 源碼的能力
- 在以下場景自動觸發查詢：
  - **LSP 報警**：編輯器報告的類型錯誤、簽名不符等
  - **編譯錯誤**：`go build` 或 `go test` 失敗時分析和修正
  - **廢棄 API**：偵測並查詢 Deprecated API 的替代品
- 查詢優先級：本機 `$GOROOT/src`（標準庫）→ `$GOPATH/pkg/mod`（已安裝依賴）→ WebFetch（遠端 pkg.go.dev）
- 返回源碼片段、API 簽名、doc comments，以及版本資訊

## Capabilities

### New Capabilities
- `go-source-query`: LLM 能查詢本機和遠端 Go 源碼，並自動在 LSP/編譯/廢棄 API 情況下觸發

### Modified Capabilities
<!-- 無現有能力受影響 -->

## Impact

- **受影響模塊**：Claude Code 的 Go 代碼編寫和除錯流程
- **新依賴**：可能需要 Go 開發環境（$GOROOT、$GOPATH）、WebFetch tool
- **使用者體驗**：無直接影響（LLM 後台自動使用，使用者無需調用）
- **性能**：主動查詢可能增加處理時間，但能提高代碼準確性
