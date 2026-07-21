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

### Requirement: 共用郵件模型支援多附件
系統 SHALL 允許共用郵件模型攜帶有序的多個附件路徑，並在同一封 multipart 郵件中為每個附件建立獨立 MIME part。既有單一附件輸入 SHALL 在本 change 保持相容。

#### Scenario: Structured send 包含多個附件
- **WHEN** 郵件模型包含兩個以上可讀附件
- **THEN** 系統將每個附件以正確 filename、content type 與 base64 encoding 寫入同一封郵件

#### Scenario: 舊單一附件輸入
- **WHEN** 既有 TUI 或呼叫端只設定單一附件相容欄位
- **THEN** 系統仍將該附件加入郵件，不要求呼叫端立即遷移

#### Scenario: 重複附件路徑
- **WHEN** 單一與多附件輸入包含相同路徑
- **THEN** 系統依首次出現順序只附加該檔案一次

### Requirement: 附件處理必須 fail-closed
系統 MUST 在 SMTP 呼叫前驗證並完整載入所有附件。任一附件不存在、不可讀、MIME type 偵測失敗或 MIME part 寫入失敗時，系統 SHALL 回傳包含附件路徑的錯誤，且不得寄出郵件。

#### Scenario: 附件不存在
- **WHEN** 任一附件路徑不存在
- **THEN** 系統回傳指出該路徑的錯誤，SMTP 呼叫次數為零

#### Scenario: 附件無法讀取
- **WHEN** 任一附件存在但無法讀取或處理
- **THEN** 系統回傳指出該路徑的錯誤，SMTP 呼叫次數為零

#### Scenario: 所有附件有效
- **WHEN** 所有附件都可成功載入並建立 MIME part
- **THEN** 系統才進入 SMTP 發送流程
