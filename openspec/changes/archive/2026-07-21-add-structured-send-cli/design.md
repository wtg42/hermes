## Context

Hermes 已有 `mail.MailCompose`、`mail.Mailer` 與 `sendmail.SMTPMailer` 的分層，也有 To/CC/BCC 驗證、multipart MIME、中文編碼與 Mailpit 整合測試；但一般單封信只能從 TUI 建立。`MailCompose` 只容納一個附件路徑，而 `SMTPMailer.buildMIMEContent` 在附件開啟或 MIME part 建立失敗時只記錄 warning，仍可能呼叫 SMTP。

新的 `hermes send` 主要供 agent 與 shell 自動化使用，因此需要比互動 TUI 更明確的 server 與安全確認邊界。本 change 與已完成但未封存的 burst safety change 平行存在；delta specs 應新增獨立 requirement，避免同時覆寫既有 requirement block。

## Goals / Non-Goals

**Goals:**

- 提供無 TUI、只靠 flags 即可執行的單封 structured send。
- 將安全預設、IPv4 server 限制與超出白名單的明確確認集中在可測試的業務層。
- 支援多 To/CC/BCC 與多附件，且所有輸入在 SMTP 副作用前完成驗證。
- 附件不存在、無法讀取或 MIME 建立失敗時 fail-closed。
- 缺少 Subject 或 Body 時產生具辨識度的隨機測試內容。
- 沿用既有 Mailer、MIME、中文編碼及 Mailpit 測試架構。

**Non-Goals:**

- 不新增 SMTP Auth、STARTTLS、implicit TLS 或任意 swaks raw arguments。
- 不新增 history、replay、last 或持久化寄信紀錄。
- 不修改 TUI 的附件選取操作；本次仍只讓 TUI 選一個檔案。
- 不將 structured send 的 IPv4 限制套用到既有 TUI、EML 或 burst command。
- 不新增獨立 HTML CLI flag，也不改變既有 multipart 文字／HTML 組裝語意。

## Decisions

### 1. `send` command 只做參數收集，安全與預設值由 sendmail service 負責

新增可由測試建立的 Cobra command factory，使用 `RunE` 收集 flags 並建立具名的 structured send options。選項交由 `sendmail` 層的 structured send service 解析、驗證、補預設內容後，再呼叫既有 `mail.Mailer`。

這讓 CLI 測試可注入假的 sender，也避免未來 history replay 或其他入口繞過安全政策。替代方案是直接在 `cmd/send_cmd.go` 組裝 `MailCompose`，但安全規則會與 Cobra 耦合，因此不採用。

### 2. Structured send 採固定安全邊界

集中定義以下常數：

- 安全寄件者：`weitingshih@rd01.softnext.com.tw`、`jllee@rd01.softnext.com.tw`、`adam@rd01.softnext.com.tw`
- 預設寄件者：`weitingshih@rd01.softnext.com.tw`
- 安全收件網域：精確的 `rd01.softnext.com.tw`
- 預設 port：`25`

未指定 From 時使用預設寄件者；未指定 To 時從安全地址中選擇一個與最終 From 不同的地址。server 沒有預設值，必須是明確的 dotted-decimal IPv4。hostname、IPv6、空值與從環境／設定推導的 server 都拒絕，且 `--confirm-outside-whitelist` 不得繞過 server 格式限制。

若寄件者不在安全寄件者清單、任一 To/CC/BCC 網域不是安全收件網域，或 port 不是 25，系統彙整原因並要求 `--confirm-outside-whitelist`。確認旗標只授權這三類安全邊界，不得略過 Email、port 格式或附件驗證。

採精確、大小寫不敏感的 mailbox／網域比較，避免子網域被誤認為安全。這比 denylist 更可預測，也與 agent 自動呼叫的風險模型一致。

### 3. 所有副作用之前建立 resolved send request

structured send pipeline 固定為：

1. 套用安全 From/To 與 port 預設值。
2. 驗證 server、port、From、To、CC、BCC。
3. 計算白名單外原因並檢查確認旗標。
4. 驗證並完整載入所有附件。
5. 只為缺少的 Subject／Body 產生隨機內容。
6. 建立 resolved `MailCompose` 並呼叫 Mailer。

任一步驟失敗都直接回傳 error，不得呼叫 Mailer 或底層 SMTP。這個 resolved request 也會成為後續 history change 可保存的穩定資料邊界，但本次不做持久化。

### 4. 多附件以新欄位擴充，暫留單一附件相容欄位

`mail.MailCompose` 新增 `Attachments []string`，既有 `Attachment string` 暫時保留。發信前會將兩者正規化成有序、去重複的附件路徑清單；TUI 可繼續只設定既有欄位，structured CLI 則使用新清單。

SMTPMailer 在寫入任何 MIME 資料前先把每個路徑載入為 `Attachment`。只要一個失敗即回傳包含路徑的 error。MIME builder 接收已載入的附件集合並逐一建立 part，任何 CreatePart／Write／Close 錯誤都向上回傳，不再記錄 warning 後繼續。

立即移除單一欄位會破壞現有 TUI 與可能的外部 Go 呼叫端；同時永久維護兩種來源會增加歧義。因此本次採相容正規化，後續 TUI 多附件 change 再遷移並評估移除舊欄位。

### 5. 隨機內容由可決定時間與 entropy 的純邏輯產生

隨機 Subject 由 emoji、中英文測試片語組合。隨機 Body 包含中英文內容、`YYYY-MM-DD HH:mm:ss +08:00` 台北時間及 8 字元十六進位 Trace-ID。實作將時間與 entropy 邊界設計為可注入或可直接傳入，使單元測試不依賴真實時鐘與不穩定亂數。

Subject 與 Body 分別補值：只缺 Subject 時不改寫使用者 Body；只缺 Body 時產生完整隨機 Body。使用者提供 Body 時不得附加時間或 Trace-ID，確保精確內容不被更動。

### 6. CLI collection flags 使用可重複／逗號分隔語意

`--to`、`--cc`、`--bcc` 與 `--attach` 使用 Cobra string-slice flags，可重複指定並接受逗號分隔。`--server` 必填；`--port` 預設 25；`--from`、`--subject`、`--body` 與 `--confirm-outside-whitelist` 為單值／布林 flags。

命令成功時回傳 nil，失敗時由 `RunE` 回傳具體錯誤並產生非零 exit status。命令不讀取 SMTP credential，不寫 history，也不進入 TUI。

### 7. 測試分層

- 純 unit tests：安全預設、地址／IPv4／port 驗證、確認原因、隨機內容格式、多附件正規化及 fail-closed。
- cmd tests：flags 解析、預設值、重複 collection flags、錯誤傳遞與 sender 呼叫次數。
- Mailpit integration：以 `127.0.0.1:1025` 執行 structured send，驗證 To/CC/BCC envelope、Subject/Body MIME 與至少兩個附件。

不在測試中連線產品 SMTP；Mailpit 為本 change 唯一環境整合目標。

## Risks / Trade-offs

- [Risk] 安全預設地址屬於組織特定設定 → 集中於單一 policy 檔並以測試鎖定，後續若需可配置化另開 change。
- [Risk] `--confirm-outside-whitelist` 可能被誤認為跳過全部驗證 → 僅在所有格式與附件驗證通過後處理白名單原因，錯誤訊息列出授權範圍。
- [Risk] `Attachment` 與 `Attachments` 暫時並存可能產生重複 → 依輸入順序正規化並去重複，文件標記單一欄位為相容用途。
- [Risk] 先讀取所有附件會增加記憶體使用 → 單封測試郵件以 fail-closed 優先；本次不處理大型串流附件。
- [Risk] 只接受 IPv4 會排除 hostname 與 IPv6 SMTP server → 限制只套用 agent-oriented `send` command，既有入口不受影響。
- [Risk] 平行 burst change 同樣觸及 CLI 與共用發信 spec → 本 change 使用獨立 ADDED requirements，避免封存順序覆蓋另一 change 的 requirement block。

## Migration Plan

1. 先以 unit tests 定義 structured options、安全政策、內容生成及附件 fail-closed。
2. 擴充 `MailCompose` 與 SMTPMailer 的多附件相容路徑，更新既有 TUI／sendmail tests。
3. 實作 structured send service 與 `hermes send` command，註冊至 root。
4. 新增 Mailpit integration cases 與 README 範例，執行 gofmt、go vet 及 `make test`。
5. 若需回復，可移除新 command/service 並保留多附件模型；沒有 DB 或外部狀態 migration。

## Open Questions

無。功能採行為對齊，不要求與其他語言工具的參數實作完全相同。
