## Why

`SendMailWithMultipart` 目前直接從 Viper map 讀值，重複執行地址驗證、附件載入、Header/MIME 組裝與 `smtp.SendMail`，與 `SMTPMailer.Send` 形成兩套容易漂移的寄送路徑；部分未檢查的 type assertion 還可能造成 panic。需要將 legacy 設定限制在資料轉換邊界，後續一律沿用既有 `MailCompose` 寄送流程。

## What Changes

- 新增 concrete legacy config-to-`MailCompose` 轉換函式，明確解析並驗證 Viper map 欄位型別。
- 將 `SendMailWithMultipart(key)` 縮減為讀取設定、轉換資料、呼叫 `SMTPMailer.Send` 與回傳既有 `(bool, error)`。
- 移除 legacy 函式內重複的地址驗證、附件載入、Header/MIME 組裝與直接 SMTP 呼叫。
- 缺少或型別錯誤的 legacy 欄位 MUST 回傳 error，不得 panic 或開始 SMTP。
- 保持函式名稱、Viper key、port 25 預設、CC header、BCC envelope、中文/MIME、附件與成功／失敗結果相容。
- 不移除 Viper、不修改 Burst、TUI、structured CLI、SMTP transport policy 或建立 adapter/config interface。

## Capabilities

### New Capabilities

- `legacy-send-adapter`: 定義 legacy Viper 設定到 `MailCompose` 的安全資料轉換，以及透過共用 SMTPMailer 寄送的等價行為。

### Modified Capabilities

無。本次不改變既有外部寄送契約。

## Impact

- 主要影響 `sendmail/use_direct_send.go` 與 legacy send tests。
- Mailpit integration tests 將驗證 legacy 入口與 SMTPMailer 的 observable message/envelope 行為一致。
- `EmailData`、Burst MIME helpers 與 `NewAttachmentLegacy` 暫時保留，留待後續獨立重構。
