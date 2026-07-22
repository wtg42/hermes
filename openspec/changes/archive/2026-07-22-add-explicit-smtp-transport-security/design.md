## Context

目前 `SMTPMailer.Send` 最終呼叫 `smtp.SendMail(addr, nil, ...)`。這個便利函式無法表達「未取得 STARTTLS 就停止」的政策，也沒有 Auth 輸入；Compose TUI 則依 Host 是否非空直接顯示 `TLS active`。Structured send 已有明確 IPv4、安全白名單、完整 preflight、history 與 replay，因此 transport 安全必須接在這些邊界之後，且不能讓密碼流入持久化資料。

本專案偏向資料導向設計。這次需要的是具體 transport 資料與一條可測試的 SMTP state sequence，不是為單一實作建立通用 transport interface。

## Goals / Non-Goals

**Goals:**

- 明確表示無 TLS／強制 STARTTLS，以及無 Auth／PLAIN Auth。
- 保留 `--server` 明確 IPv4 作為 TCP 連線目標，另以 TLS server name 驗證憑證身分。
- 確保 TLS、憑證或 Auth 失敗時，SMTP envelope 與 message bytes 都尚未送出。
- 密碼只存在於當次程序記憶體，不進入 history、輸出或 replay record。
- 以 concrete structs、枚舉值與函式組合實作，避免不必要的介面階層。

**Non-Goals:**

- implicit TLS（通常為 port 465）、OAuth、client certificate 或其他 SASL mechanism。
- 在 Compose TUI 新增 TLS/Auth 輸入欄位。
- 改變既有收件網域、From、port 或明確 IPv4安全政策。
- 建立可插拔 SMTP provider／transport framework。

## Decisions

### 1. 郵件資料與 transport 資料分離

新增 concrete `SMTPTransportConfig` 類型，包含 connect server、port、TLS mode、TLS server name、Auth mode、username 與當次 password。`MailCompose` 繼續描述 message/envelope；MIME builder 不接觸 transport secret。

選擇 concrete struct 而不是 `Transport` interface，因為目前只有 SMTP 一種傳輸，測試可透過現有函式注入點或低階 dial hook 隔離。等到存在第二種真實 transport 再評估抽象。

### 2. CLI 使用封閉集合參數，密碼只從 stdin 讀取

`hermes send` 新增：

- `--tls-mode none|required`
- `--tls-server-name <name>`
- `--auth-mode none|plain`
- `--auth-username <value>`
- `--auth-password-stdin`

預設均為 `none`，保留既有未加密、未認證流程的相容性。PLAIN Auth 必須搭配 `tls-mode=required`、username 與 `--auth-password-stdin`；命令從 stdin 讀取一行密碼，去除行尾，不提供會暴露於 process list 的 password value flag。

未選用環境變數是為了避免秘密被子程序繼承；未選用互動 prompt 是為了維持 structured CLI 可由 agent／script 自動呼叫。

### 3. 強制 STARTTLS 使用明確 SMTP sequence

安全路徑依序執行 TCP dial、讀取 greeting、EHLO、確認 STARTTLS extension、`StartTLS`（啟用標準憑證驗證與指定 ServerName）、再次取得 server capabilities、執行 Auth，最後才送出 MAIL/RCPT/DATA。任一步驟錯誤即關閉連線並回傳階段化錯誤。

`tls-mode=required` 不允許降級；server 未宣告 STARTTLS 也視為失敗。`tls-mode=none` 沿用明文 SMTP，但不得搭配 PLAIN Auth，避免送出可還原的 credential。

### 4. Preflight 先驗證 transport 組合，再讀取密碼或連線

Transport 參數的枚舉值、必要搭配、TLS server name、Auth username 與 stdin 模式，必須和既有地址、附件、安全確認一起在 SMTP 前驗證。只有所有靜態 preflight 通過後才讀取 password；password 為空時仍在 dial 前失敗。

白名單確認只授權既有 From／收件人／port 邊界，不能略過 transport 驗證。

### 5. History 保存可重建政策，不保存秘密

History request schema 升版，保存 TLS mode、TLS server name、Auth mode 與 username，但不保存 password、stdin 內容或「已認證」狀態。Replay 仍走目前 structured pipeline；需要 PLAIN Auth 的 record 必須由本次 replay 再提供 `--auth-password-stdin`，否則在連線前失敗。

舊 schema record 按既有版本規則處理，不進行隱式 migration；這避免錯誤解讀安全預設。實作時以新版本常數與明確驗證更新測試 fixture。

### 6. TUI 只顯示設定狀態，不宣稱連線結果

Compose TUI 本次沒有 transport negotiation state，因此 Host/Port 填入後只顯示 SMTP target（例如 `SMTP target smtp.example.com:587`）。只有未來 model 真正收到已驗證連線結果時，才可顯示 TLS active。

### 7. 測試採協定狀態與既有整合測試分層

使用本機 ephemeral SMTP test server 與測試 CA/certificate 驗證 STARTTLS、錯誤憑證、缺少 extension、Auth 成功／失敗，以及失敗前沒有 MAIL/RCPT/DATA。敏感資料 redaction 與 history schema 使用 unit tests；既有 Mailpit 繼續驗證明文預設路徑、MIME 與實際投遞。

## Risks / Trade-offs

- [標準函式庫 SMTP API 低階且功能有限] → 將協定流程限制在 STARTTLS + PLAIN，先以小型具體函式完成；只有測試證明維護成本不合理時才提出依賴變更。
- [PLAIN Auth 誤用於明文連線會洩漏密碼] → preflight 強制 PLAIN 只能搭配 required STARTTLS，且 Auth 必須在 TLS 成功後才執行。
- [IPv4 connect target 與憑證名稱不同] → 保留兩個明確欄位，TLS 驗證只使用 `tls-server-name`，不做 DNS 解析來替換 connect target。
- [history schema 升版讓既有紀錄暫時不可讀] → 保持 fail-closed 並在錯誤中指出版本；不以猜測方式補上 transport policy。
- [stdin 密碼讓 shell pipeline 稍複雜] → 提供清楚的非互動用法，但不以便利性換取 process-list 暴露。
- [錯誤字串意外包含 Auth 值] → 錯誤只標示 transport 階段與 server 回應摘要，不格式化完整 config 或 credential。
