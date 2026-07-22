## MODIFIED Requirements

### Requirement: 統一的郵件發送 API
系統 SHALL 提供共用的郵件構建與發送邏輯，用於所有發送場景（CLI、TUI、Burst mode），支援純文字、HTML 內容及附件；Burst mode SHALL 依預先驗證的設定使用固定或隨機 From 與 To。

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
- **WHEN** 用戶啟動 Burst mode，指定數量、SMTP 伺服器及通過驗證的固定／隨機地址設定
- **THEN** 系統使用共用的郵件構建邏輯，依設定併發發送測試郵件
