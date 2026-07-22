## Why

Structured send 目前以 `*SMTPMailer` concrete type assertion 決定是否套用 `SMTPTransportConfig`，使正式 SMTPMailer 與測試／替代 Mailer 走不同執行路徑，transport policy 可能被靜默忽略。需要把已解析郵件與 transport 組成單一 concrete plan，讓驗證、寄送結果與 history 都以同一份資料為準。

## What Changes

- 新增 concrete `StructuredSendPlan`，同時保存 resolved message 與 validated SMTP transport data。
- 將 structured send 明確分成「建立完整 plan」與「執行 plan」兩個資料流階段。
- 移除依賴 `*SMTPMailer` concrete type assertion 的 transport／legacy 雙路徑；執行函式永遠接收完整 plan。
- 讓 history record 直接由 execution 使用的 plan 與 result 建立，不再從 `MailCompose` 反推 transport metadata。
- 以小型函式值注入單次 plan attempt 供測試使用，不新增 Repository、Service、Factory 或通用 transport interface。
- 保持所有 CLI flags、預設值、安全驗證、SMTP/TLS/Auth、MIME、history schema、replay 與 TUI 行為相容。

## Capabilities

### New Capabilities

- `structured-send-plan`: 定義 structured send 在任何副作用前建立完整、可驗證 plan，並以同一 plan 執行 SMTP 與建立 history 的一致性要求。

### Modified Capabilities

無。本次為內部執行資料流重構，不改變既有 capability 的外部契約。

## Impact

- 主要影響 `sendmail/structured_send.go`、`sendmail/history.go` 與其測試。
- `cmd/send_cmd.go`、history replay 與 integration tests 需確認仍透過相同 structured pipeline。
- `mail.Mailer` 可保留供 TUI／EML 使用，但 structured send 不再透過 type assertion 選擇 transport 行為。
- 不變更 CLI、history JSON schema、SMTP protocol、TUI、Burst 或 legacy Viper 寄送入口。
