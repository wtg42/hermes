## 1. Transport 資料與 preflight

- [x] 1.1 先新增 table-driven failing tests，覆蓋 TLS/Auth mode、必要欄位、無作用參數、PLAIN-over-plaintext 與明確 IPv4/TLS server name 分工
- [x] 1.2 新增 concrete SMTP transport config 與封閉集合常數，不建立通用 transport interface
- [x] 1.3 實作 transport preflight，確保所有靜態錯誤在 password stdin、Mailer 與 SMTP 前失敗
- [x] 1.4 執行 transport/preflight 單元測試並完成必要重構

## 2. SMTP STARTTLS 與 Auth 流程

- [x] 2.1 建立本機 ephemeral SMTP test server 與測試憑證 fixture，記錄 EHLO、STARTTLS、AUTH、MAIL、RCPT、DATA 的事件順序
- [x] 2.2 先新增 failing tests，覆蓋 required STARTTLS 成功、缺少 extension、握手／憑證失敗且不得降級
- [x] 2.3 實作 TCP、SMTP greeting/EHLO、required STARTTLS 與標準 certificate/server-name 驗證流程
- [x] 2.4 先新增 failing tests，覆蓋 PLAIN Auth 成功、拒絕與「TLS 成功前不得 Auth／envelope」
- [x] 2.5 實作 TLS 後 PLAIN Auth 與 MAIL/RCPT/DATA 流程，確保每個階段失敗即關閉連線
- [x] 2.6 新增錯誤階段與 password redaction tests，確認任何輸出與錯誤鏈都不洩漏 secret
- [x] 2.7 執行 SMTP transport 單元／本機協定測試並完成必要重構

## 3. Structured CLI 整合

- [x] 3.1 先新增 failing command tests，覆蓋 transport flags 預設值、完整參數與所有無效組合
- [x] 3.2 為 `hermes send` 加入 TLS/Auth flags，解析為 concrete transport config 並保持既有明文無 Auth 預設
- [x] 3.3 先新增 failing stdin tests，覆蓋靜態 preflight 後讀取一行 password、空值與未要求 stdin 的路徑
- [x] 3.4 實作 `--auth-password-stdin`，移除行尾且不增加接受 password value 的 CLI flag
- [x] 3.5 將 structured send coordinator 與 SMTPMailer 接到新 transport 流程，驗證失敗時 Mailer／SMTP 呼叫次數為零
- [x] 3.6 執行 cmd 與 structured send 單元測試並完成必要重構

## 4. History schema 與 replay

- [x] 4.1 先新增 failing history tests，覆蓋新版 transport metadata、舊版本拒絕及 password／secret 永不序列化
- [x] 4.2 升級 history schema，保存 TLS/Auth policy 與 username，並讓結果錯誤保持 secret-safe
- [x] 4.3 先新增 failing replay tests，覆蓋 authenticated record 必須重新取得 password、preflight 先於 stdin 及 `--no-history`
- [x] 4.4 擴充 history replay flags 與 coordinator，確保 record 不提供 password 或可重用認證狀態
- [x] 4.5 執行 history list/show/replay 單元與整合測試並完成必要重構

## 5. TUI 狀態正確性

- [x] 5.1 先新增 Compose View failing tests，證明只填 Host/Port 時不得顯示 Connected、TLS active 或 Auth 成功
- [x] 5.2 將底部狀態改為即時顯示 `SMTP target <host>:<port>`，且不改變既有 Compose 寄送流程
- [x] 5.3 執行 Compose TUI View／Update 相關測試

## 6. 文件與整體驗證

- [x] 6.1 更新 README structured send 參數、安全限制、stdin password 範例、STARTTLS 行為與本次不支援項目
- [x] 6.2 執行 `gofmt`、`go vet ./...` 與受影響套件單元測試
- [x] 6.3 執行 `make test`，確認既有 Mailpit MIME／投遞、structured send、history 與 burst 行為無回歸
- [x] 6.4 執行 OpenSpec strict validation，並核對實作與所有 transport security scenarios 一致
