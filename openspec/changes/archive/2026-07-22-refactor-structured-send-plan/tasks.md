## 1. Characterization 與 plan 資料

- [x] 1.1 先新增 failing characterization tests，證明 TLS/Auth transport 在非 `*SMTPMailer` 測試替身路徑會被忽略
- [x] 1.2 新增 `StructuredSendPlan` concrete data type 與 plan-aware attempt function type，不新增 interface 或 factory
- [x] 1.3 新增 plan shape tests，覆蓋 resolved message、附件順序、server/port、TLS/Auth 與當次 credential

## 2. Plan builder

- [x] 2.1 先新增 table-driven failing tests，鎖定 defaults、驗證錯誤順序、隨機內容與所有 preflight failure 的零 attempt 副作用
- [x] 2.2 將 structured options 的 resolve、安全評估、transport ready validation、附件 preflight 與隨機內容收斂為 plan build 流程
- [x] 2.3 確保 plan 完成後不再修改 Compose slices、attachments 或 transport data，必要時建立防止 aliasing 的資料副本
- [x] 2.4 執行 plan builder 與既有 structured send 單元測試並完成必要重構

## 3. 單一路徑 execution

- [x] 3.1 先新增 failing tests，要求任何 plan-aware attempt 都收到完整 plan，且 attempt error 只執行一次
- [x] 3.2 將 `StructuredSender` 改為持有單一 plan-aware attempt，移除 `*SMTPMailer` concrete type assertion 與 MailCompose-only fallback
- [x] 3.3 更新正式 structured command，明確以 `SMTPMailer.SendWithTransport(plan.Compose, plan.Transport)` 執行 plan
- [x] 3.4 更新 structured send 測試替身與 call sites，讓測試直接記錄 plan 而非依賴 `mail.Mailer` concrete type
- [x] 3.5 使用 `rg` 確認 structured execution 不再依 mailer concrete type 分流，並執行 cmd/sendmail 相關測試

## 4. History 與 replay 一致性

- [x] 4.1 先新增 failing tests，證明 history 的 message 與 transport metadata 必須等於 attempt 收到的同一 plan
- [x] 4.2 將 history record constructor 改為接受 `StructuredSendPlan`，移除 compose 加 optional variadic transport 的入口
- [x] 4.3 更新 structured history coordinator，僅在 attempt 後從 execution plan 建立成功或失敗 record
- [x] 4.4 更新 replay、`--no-history`、send/history 雙錯誤與 password redaction tests，確認每條路徑最多 attempt 一次
- [x] 4.5 執行 history unit tests 與 Mailpit send/replay integration tests，確認 history v2 JSON schema 不變

## 5. 相容性與驗證

- [x] 5.1 更新公開註解與 README（若內部 Go API 用法有需要），不新增外部 CLI 行為或新文件
- [x] 5.2 執行 `gofmt`、`go vet ./...` 與 `go test ./... -race -cover`
- [x] 5.3 執行 `make test`，驗證 Mailpit MIME、structured send、history replay、Burst 與 TUI 無回歸
- [x] 5.4 執行 OpenSpec strict validation 與 `git diff --check`，核對實作未觸及 TUI、Burst、legacy Viper 或 history schema 範圍
