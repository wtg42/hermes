## 1. Sendmail 安全行為測試（Red）

- [x] 1.1 在 `sendmail/burst_send_test.go` 新增表格測試，覆蓋 From/To 皆固定、僅 From 固定、僅 To 固定、兩者皆隨機，並驗證固定值同時出現在 SMTP envelope 與郵件 header。
- [x] 1.2 新增網域安全測試，覆蓋內建 `rd01.softnext.com.tw`、未授權非表列網域、精確 `--allow-domain` 授權、父網域不涵蓋子網域，以及多網域中任一未授權即失敗。
- [x] 1.3 新增預先驗證失敗測試，覆蓋缺少隨機網域、無效固定地址、無效 domain／allow-domain，並斷言每種錯誤的 `SendMail` 呼叫次數皆為零。
- [x] 1.4 先執行 `go test ./sendmail`，確認新增測試能重現尚未實作的需求。

## 2. Sendmail 地址計畫與發送實作（Green）

- [x] 2.1 新增具名 burst 選項模型與可測試的預先解析結果，取代持續擴張的位置參數，並讓 `BurstModeSendMail` 回傳 `error`。
- [x] 2.2 實作單一固定 mailbox address、隨機 domain 清單與 allow-domain 的格式解析；網域比較前 trim 並轉小寫，但保留 mailbox local part。
- [x] 2.3 實作內建安全網域 `rd01.softnext.com.tw`、所有實際網域的精確授權檢查及彙整錯誤，確保通過前不建立 goroutine 或呼叫 SMTP。
- [x] 2.4 依預先驗證結果實作四種 From/To 組合，固定地址同步用於 header 與 envelope，未指定地址維持既有隨機池、主旨、內容及併發行為。
- [x] 2.5 更新 `sendmail` 內既有測試與呼叫端以使用新選項模型，執行 `go test ./sendmail` 確認通過。

## 3. Burst CLI 測試與旗標（Red → Green）

- [x] 3.1 先在 `cmd/burst_cmd_test.go` 新增失敗測試：只有 quantity/host/port 時回傳缺少 domain 錯誤且不寄信，非表列固定或隨機網域未授權時亦不寄信。
- [x] 3.2 新增成功測試：安全隨機 domain、From/To 皆固定，以及非表列網域搭配精確且可重複的 allow-domain 均能把正確選項傳入發送流程。
- [x] 3.3 在 `cmd/burst_cmd.go` 註冊並綁定 `--domain`、`--from`、`--to` 與可重複的 `--allow-domain`，將命令改為 `RunE` 並回傳解析、驗證與發送錯誤。
- [x] 3.4 執行 `go test ./cmd`，確認 CLI 成功與零寄送錯誤路徑全部通過。

## 4. 文件與完整驗證

- [x] 4.1 更新 `README.md` 的 Burst mode 參數、breaking migration、四種固定／隨機範例、安全網域及 `--allow-domain` 風險說明。
- [x] 4.2 執行 `gofmt`／`go fmt` 格式化受影響 Go 檔案，並執行 `go vet ./...`。
- [x] 4.3 執行 `make test`（`go test ./... -race -cover`），確認全套測試及 race detector 通過。
