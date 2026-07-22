## 1. Legacy 設定 characterization

- [x] 1.1 先新增 table-driven tests，覆蓋完整設定、port 25 預設及 cc/bcc/attachment 可選欄位
- [x] 1.2 新增缺少或錯型別 host/from/to/subject/contents 的 failing tests，確認不得 panic 或呼叫 SMTP
- [x] 1.3 鎖定既有 comma-separated To/CC/BCC、空值與錯誤欄位語意

## 2. Concrete adapter

- [x] 2.1 實作未匯出的 legacy map-to-MailCompose 純資料轉換函式，不新增 interface/factory
- [x] 2.2 確保 converter 不讀 Viper、不載入附件、不驗證 SMTP 地址且沒有 I/O 副作用
- [x] 2.3 執行 converter 單元測試並完成必要重構

## 3. 共用 SMTPMailer pipeline

- [x] 3.1 先新增 failing test，證明 legacy 成功／失敗皆只呼叫共用 Mailer pipeline 一次
- [x] 3.2 將 `SendMailWithMultipart` 改為讀取 map、轉換 MailCompose、呼叫 `SMTPMailer.Send` 與映射 `(bool, error)`
- [x] 3.3 移除 legacy 函式內重複的地址驗證、附件、Header/MIME、envelope 與直接 SendMail 流程
- [x] 3.4 使用 `rg` 確認 `SendMailWithMultipart` 不再直接組裝 MIME 或呼叫 SendMail，且 Burst helpers 保留

## 4. Observable 相容性

- [x] 4.1 更新既有 unit tests，覆蓋 To/CC visible headers、BCC-only envelope 與 invalid address fail-closed
- [x] 4.2 驗證中文 Subject/Contents、plain/HTML multipart、附件與 SMTP error 結果相容
- [x] 4.3 執行 Mailpit legacy integration tests，確認 message/envelope observable behavior 無回歸

## 5. 整體驗證

- [x] 5.1 更新公開註解（如需要），不增加外部功能或新文件
- [x] 5.2 執行 `gofmt`、`go vet ./...` 與 `go test ./... -race -cover`
- [x] 5.3 執行 `make test`，確認 Burst、TUI、structured send、history 與 Mailpit 無回歸
- [x] 5.4 執行 OpenSpec strict validation、`git diff --check` 與 scope audit，確認未修改 Burst/TUI/structured/history 行為
