## 1. 多附件與 fail-closed（Red → Green）

- [x] 1.1 在 `sendmail` 測試新增 `MailCompose.Attachments` 多附件案例，驗證兩個以上附件各自具有正確 filename、content type、base64 內容，並覆蓋重複路徑去重。
- [x] 1.2 新增相容與錯誤案例：既有單一 `Attachment` 仍可寄送；任一附件不存在、不可讀或處理失敗時 `SendMail` 呼叫次數為零且 error 包含路徑。
- [x] 1.3 執行 `go test ./sendmail`，確認新附件測試先以現有單附件／warning 行為失敗。
- [x] 1.4 在 `mail.MailCompose` 新增 `Attachments []string` 並保留單一 `Attachment` 相容欄位，實作有序去重、SMTP 前完整載入與多 MIME parts，移除附件失敗後繼續寄送的 warning 路徑。
- [x] 1.5 更新既有 TUI／sendmail 測試與呼叫點，執行 `go test ./sendmail ./tui` 確認多附件、fail-closed 與單附件相容行為通過。

## 2. Structured send 內容與安全核心（Red → Green）

- [x] 2.1 新增可決定時間與 entropy 的隨機內容測試，驗證 emoji／中英文 Subject、`+08:00` 台北時間、8 字元 Trace-ID，以及只補缺少欄位而不改寫使用者 Body。
- [x] 2.2 新增安全預設測試，驗證預設 From、不同的安全 To、port 25，以及指定任一安全 From 時預設 To 仍不相同。
- [x] 2.3 新增 server／格式測試，覆蓋有效 dotted IPv4、缺少 server、hostname、localhost、IPv6、無效 Email 與無效 port，並斷言確認旗標不能繞過基本驗證。
- [x] 2.4 新增白名單測試，覆蓋三個安全寄件者、精確安全收件網域、子網域、外部 To/CC/BCC、非 25 port、多原因彙整與 `--confirm-outside-whitelist` 成功授權。
- [x] 2.5 新增零副作用測試，驗證任一 structured preflight 或附件驗證失敗時注入的 Mailer 呼叫次數為零，並先執行 focused tests 取得 Red 結果。
- [x] 2.6 實作具名 structured send options、resolved request、隨機內容 generator、安全預設、IPv4／port／地址驗證及白名單確認政策。
- [x] 2.7 實作 structured send service 的固定 pipeline，於全部驗證及附件預載完成後才建立 `MailCompose` 並呼叫注入的 Mailer。
- [x] 2.8 執行 structured send focused unit tests，確認預設、內容、安全確認與零副作用案例全部通過。

## 3. `hermes send` CLI（Red → Green）

- [x] 3.1 在 `cmd` 新增測試，覆蓋最小 `--server` 呼叫、完整 flags、重複與逗號分隔 To/CC/BCC/Attach，並驗證輸入順序及 confirmation flag 傳入 service。
- [x] 3.2 新增 CLI 錯誤與 help 測試，確認 service error 由 `RunE` 傳回、缺少 server 不寄信，且 `hermes send --help` 不啟動 TUI 或呼叫 service。
- [x] 3.3 新增可測試的 send command factory，註冊 `--server`、`--port`、`--from`、`--to`、`--cc`、`--bcc`、`--subject`、`--body`、`--attach`、`--confirm-outside-whitelist`，並加入 root command。
- [x] 3.4 執行 `go test ./cmd`，確認 structured collection flags、錯誤傳遞、help 與 root routing 全部通過。

## 4. Mailpit 整合、文件與完整驗證

- [x] 4.1 新增 Mailpit integration test，透過 structured send 將 To/CC/BCC、中文與 emoji Subject/Body 送至 `127.0.0.1:1025`，驗證 envelope、可見 headers 與 MIME 內容。
- [x] 4.2 新增 structured 多附件 Mailpit case，驗證至少兩個附件的 filename／內容；另驗證不存在附件時 Mailpit 郵件數量不增加。
- [x] 4.3 更新 `README.md`，說明 `hermes send` 最小／完整範例、安全預設、IPv4 限制、多值 flags、多附件 fail-closed 與 `--confirm-outside-whitelist` 風險。
- [x] 4.4 對受影響 Go 檔執行 gofmt，並執行 `go vet ./...`。
- [x] 4.5 執行 `make test`，確認 unit、Mailpit integration、MIME assertions、coverage 與 race detector 全部通過。
