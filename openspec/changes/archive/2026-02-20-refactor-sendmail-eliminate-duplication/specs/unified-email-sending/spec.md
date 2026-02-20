## MODIFIED Requirements

### Requirement: 統一的郵件發送 API
系統 SHALL 提供單一的郵件發送函數（`SendMailWithMultipart()`），用於所有發送場景（CLI、TUI、Burst mode），支援純文字、HTML 內容及附件。

#### Scenario: 發送純文字郵件
- **WHEN** 用戶通過 TUI 選擇純文字模式並填入寄件人、收件人、主題、內容
- **THEN** 系統使用 SendMailWithMultipart() 發送郵件，郵件內容以 base64 編碼，使用 multipart/mixed 格式

#### Scenario: 發送 HTML 郵件
- **WHEN** 用戶通過 TUI 選擇 HTML 模式並填入相應內容
- **THEN** 系統使用 SendMailWithMultipart() 發送郵件，包含 text/plain 和 text/html 兩個 MIME 部分

#### Scenario: 發送帶附件的郵件
- **WHEN** 用戶通過 TUI 上傳附件
- **THEN** 系統使用 SendMailWithMultipart() 發送郵件，包含文字內容和附件 MIME 部分

#### Scenario: Burst mode 發送多封郵件
- **WHEN** 用戶啟動 Burst mode 指定數量和 SMTP 伺服器
- **THEN** 系統使用共用的郵件構建邏輯，併發發送隨機生成的測試郵件

### Requirement: 一致的錯誤處理
系統 SHALL 統一使用 `error` 返回值進行錯誤處理，避免 panic 或忽略錯誤。所有郵件發送相關函數 SHALL 返回 `(bool, error)` 或類似的結構。

#### Scenario: SMTP 連接失敗
- **WHEN** SMTP 伺服器無法連接
- **THEN** 函數返回 error，調用者可適當処理（如記錄日誌、顯示錯誤訊息）

#### Scenario: 無效的電子郵件地址
- **WHEN** 收件人地址不符合 email 格式
- **THEN** 系統驗證並過濾無效地址，返回 error 表示有無效地址被排除

### Requirement: 可配置的 SMTP 參數
系統 SHALL 允許通過 TUI 配置 SMTP 伺服器、埠口、發件人、收件人等參數。埠口預設值為 25。

#### Scenario: 修改 SMTP 埠口
- **WHEN** 用戶在 TUI 中修改埠口值
- **THEN** 系統記住此設定，用於後續郵件發送

#### Scenario: 使用預設埠口
- **WHEN** 用戶未指定埠口值
- **THEN** 系統使用預設值 25

### Requirement: 中文編碼支援
系統 SHALL 支援中文主題和內容，使用 RFC 2047 base64 編碼處理主題，內容部分使用 base64 編碼。

#### Scenario: 發送中文主題郵件
- **WHEN** 用戶在主題中輸入中文
- **THEN** 系統自動編碼為 =?UTF-8?B?...?= 格式，郵件客戶端可正確顯示

#### Scenario: 發送中文內容郵件
- **WHEN** 用戶在內容中輸入中文
- **THEN** 系統使用 base64 編碼內容，收件人郵件客戶端可正確顯示
