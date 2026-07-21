## ADDED Requirements

### Requirement: Mailpit 驗證 structured send
Mailpit integration tests SHALL 經 structured send service 實際發送郵件至 `127.0.0.1:1025`，並透過 Mailpit API 驗證收件 envelope、MIME 內容與多附件。

#### Scenario: Structured send 到 Mailpit
- **WHEN** integration test 使用明確確認，將包含 To、CC、BCC、中文 Subject 與 Body 的 structured message 發送至 `127.0.0.1:1025`
- **THEN** Mailpit 成功接收郵件，To/CC header 與所有 SMTP envelope recipients 正確，BCC 不出現在可見 header

#### Scenario: Structured send 多附件到 Mailpit
- **WHEN** integration test 透過 structured send 附加至少兩個不同檔案
- **THEN** Mailpit 解析出兩個附件，且 filename 與內容符合測試輸入

#### Scenario: Structured send 附件失敗不投遞
- **WHEN** integration test 提供至少一個不存在附件
- **THEN** structured send 在 SMTP 前失敗，Mailpit 郵件數量不增加
