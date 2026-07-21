## Why

Hermes 的一般單封寄信目前主要依賴 TUI，缺少適合 agent 與 shell 自動化呼叫的 structured CLI；現有單附件模型也會在附件處理失敗時記錄警告後繼續寄信。新增安全、可預測的 `hermes send`，可補齊自動化寄信能力，並先建立後續 Auth/TLS、history 與 TUI 共用的發信核心。

## What Changes

- 新增 `hermes send` structured CLI，支援明確 SMTP IPv4 server、port、From、多 To/CC/BCC、Subject、Body 及可重複附件參數。
- 未指定 From/To 時套用安全預設地址，且預設寄件者與收件者不得相同。
- 未指定 Subject 或 Body 時，分別產生包含中英文、emoji、台北時間與 Trace-ID 的測試內容；使用者提供 Body 時不附加自動時間或 Trace-ID。
- 新增一般單封寄信安全政策：server 必須是明確 IPv4；寄件者白名單、收件網域與 port 超出安全範圍時，必須提供 `--confirm-outside-whitelist`。
- 將郵件資料模型與 MIME 組裝擴充為多附件；所有附件必須在 SMTP 呼叫前成功驗證與載入，任一附件不存在或處理失敗時整封拒絕。
- 保留既有 To/CC/BCC 驗證、BCC envelope 行為、中文編碼與 multipart MIME 能力。
- 新增 structured CLI 的單元測試及 Mailpit SMTP/MIME 整合測試，並更新 README。
- 本次不新增 SMTP Auth/TLS、history/replay、TUI 多附件操作或任意 swaks raw arguments。

## Capabilities

### New Capabilities

- `structured-send-cli`: 定義 `hermes send` 的 flags、安全預設、多收件人、多附件、隨機內容與命令結果。
- `direct-send-safety`: 定義 agent-friendly 單封寄信的 IPv4、寄件者白名單、收件網域、port 與明確確認邊界。

### Modified Capabilities

- `cli-mode-routing`: 新增不啟動 TUI、直接執行單封寄信的 `send` 子命令。
- `unified-email-sending`: 將共用郵件模型與 MIME 組裝擴充為多附件，並改為附件錯誤時 fail-closed。
- `mailpit-integration-testing`: 增加 structured CLI 經真實 SMTP 送入 Mailpit 並驗證 envelope、MIME 與多附件的整合情境。

## Impact

- CLI：新增 `cmd/send_cmd.go` 與對應測試，並註冊至 root command。
- Domain model：`mail.MailCompose` 增加多附件表示；為降低相容性風險，既有單一 `Attachment` 欄位暫時保留並正規化至附件清單。
- 發信邏輯：`sendmail.SMTPMailer`、MIME builder、附件載入及安全／預設內容服務。
- TUI：外部操作行為不變，既有單附件值仍可透過相容欄位發送。
- 測試與文件：新增 cmd/sendmail 單元測試、Mailpit integration cases 與 structured CLI 使用說明。
- 不新增第三方 runtime dependency，不進行實際產品 SMTP 驗收。
