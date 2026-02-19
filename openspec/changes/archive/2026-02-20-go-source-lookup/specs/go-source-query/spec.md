## ADDED Requirements

### Requirement: LLM 能自動查詢 Go 標準庫源碼
LLM 在遇到標準庫相關的 LSP 報警、編譯錯誤或廢棄 API 警告時，能查詢本機 Go 標準庫源碼，取得函數簽名、doc comments 和相關源碼片段。

#### Scenario: 標準庫函數簽名查詢
- **WHEN** LLM 遇到標準庫函數的 LSP 報警（如參數數量錯誤）
- **THEN** 自動查詢 $GOROOT/src 中該函數的定義，返回準確簽名和 doc comments

#### Scenario: 標準庫廢棄 API 查詢
- **WHEN** LLM 偵測到代碼使用了標記為 Deprecated 的標準庫 API
- **THEN** 查詢該 API 的 doc comments，提取廢棄原因和推薦替代品

#### Scenario: 標準庫實現細節查詢
- **WHEN** LLM 需要理解標準庫內部實現（如 encoding/json 的序列化邏輯）
- **THEN** 查詢源碼並返回核心函數片段和相關類型定義

### Requirement: LLM 能查詢第三方依賴套件源碼
LLM 在編譯錯誤或 LSP 報警涉及第三方套件時，能查詢本機已安裝的依賴（基於 go.mod），或從遠端源查詢。

#### Scenario: 本機已安裝依賴的查詢
- **WHEN** LLM 需要查詢 go.mod 中列出的第三方套件
- **THEN** 查詢 $GOPATH/pkg/mod 中對應版本的源碼，返回函數簽名和 doc comments

#### Scenario: 遠端依賴查詢
- **WHEN** 本機 $GOPATH/pkg/mod 中找不到該套件，但需要查詢
- **THEN** 使用 WebFetch 查詢 pkg.go.dev，返回該套件的 API 文檔或源碼片段

#### Scenario: 版本資訊返回
- **WHEN** 查詢第三方套件時
- **THEN** 返回結果中包含當前查詢的版本（go.mod 中的版本或遠端最新版本）

### Requirement: 自動觸發查詢
LLM 在以下三種情況自動觸發查詢，無需使用者手動調用。

#### Scenario: LSP 報警自動觸發
- **WHEN** Claude Code 偵測到編輯器 LSP 報警（如 undefined variable, wrong argument count）
- **THEN** 自動查詢相關套件源碼，在返回給使用者前用查詢結果驗證或修正代碼

#### Scenario: 編譯錯誤自動觸發
- **WHEN** `go build` 或 `go test` 執行失敗
- **THEN** 解析編譯錯誤訊息，自動查詢涉及的套件和函數，提供修正建議

#### Scenario: 廢棄 API 自動觸發
- **WHEN** LLM 偵測到代碼中使用了標記為 Deprecated 的 API（通過源碼掃描或 doc comments）
- **THEN** 自動查詢該 API 的棄用資訊和替代品

### Requirement: 查詢優先級和資訊完整性
查詢遵循明確的優先級順序，返回信息包含必要的上下文。

#### Scenario: 本機優先查詢
- **WHEN** LLM 需要查詢函數
- **THEN** 優先在 $GOROOT/src（標準庫）和 $GOPATH/pkg/mod（依賴）中查詢，只有本機找不到時才查遠端

#### Scenario: 返回信息包含簽名和文檔
- **WHEN** 查詢成功
- **THEN** 返回結果包含：函數簽名、doc comments、相關類型定義、版本資訊

#### Scenario: 查詢失敗降級
- **WHEN** 無論本機還是遠端都查詢失敗
- **THEN** 返回查詢失敗訊息，LLM 依賴訓練知識進行

### Requirement: 性能和避免重複查詢
實現簡單的去重和快取機制，避免同一查詢短時間內重複觸發。

#### Scenario: 查詢去重
- **WHEN** 同一個 LSP 報警或編譯錯誤在短時間內（15 分鐘）多次出現
- **THEN** 第一次查詢後快取結果，後續查詢返回快取而不重新查詢

#### Scenario: 快取清理
- **WHEN** 工作階段結束或使用者手動清理
- **THEN** 清除快取以便下次使用者改動代碼後能重新查詢
