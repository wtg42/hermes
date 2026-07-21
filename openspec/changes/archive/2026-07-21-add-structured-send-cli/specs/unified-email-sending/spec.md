## ADDED Requirements

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
