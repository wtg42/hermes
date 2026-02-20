## MODIFIED Requirements

### Requirement: 一致的錯誤處理
系統 SHALL 統一使用 `error` 返回值進行錯誤處理，避免 panic 或忽略錯誤。所有郵件發送相關函數 SHALL 返回 `(bool, error)` 或類似的結構。當任何收件人字段（To、Cc、Bcc）包含無效地址時，系統 SHALL 停止郵件發送並返回包含具體無效地址和字段名稱的 error。

#### Scenario: SMTP 連接失敗
- **WHEN** SMTP 伺服器無法連接
- **THEN** 函數返回 error，調用者可適當処理（如記錄日誌、顯示錯誤訊息）

#### Scenario: To 字段包含無效地址
- **WHEN** To 字段中有一或多個不符合 email 格式的地址
- **THEN** 系統返回 error，訊息清楚指出"To"字段及無效地址清單，不發送郵件

#### Scenario: Cc 字段包含無效地址
- **WHEN** Cc 字段中有一或多個不符合 email 格式的地址
- **THEN** 系統返回 error，訊息清楚指出"Cc"字段及無效地址清單，不發送郵件

#### Scenario: Bcc 字段包含無效地址
- **WHEN** Bcc 字段中有一或多個不符合 email 格式的地址
- **THEN** 系統返回 error，訊息清楚指出"Bcc"字段及無效地址清單，不發送郵件

#### Scenario: 所有收件人字段都有效
- **WHEN** To、Cc、Bcc 中的所有地址都符合 email 格式（或該字段為空）
- **THEN** 系統繼續郵件發送流程

#### Scenario: 使用者在 TUI 中看到驗證錯誤
- **WHEN** 使用者在 TUI 編輯區輸入無效的 Cc/Bcc 並嘗試發送
- **THEN** 系統在編輯區顯示具體的錯誤訊息，讓使用者直接修正並重新發送，無需離開編輯界面
