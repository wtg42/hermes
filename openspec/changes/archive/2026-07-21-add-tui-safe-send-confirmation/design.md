## Context

目前 `StructuredSender` 在 `sendmail/structured_send.go` 內自行驗證三個安全寄件者、精確安全收件網域與 port 25；Compose TUI 則在 `Ctrl+S` 後直接建立 `mail.MailCompose` 並非同步呼叫注入的 `mail.Mailer`。因此同一封郵件從 structured CLI 寄送時有安全確認，從 TUI 寄送時卻沒有。

本變更跨越 `sendmail` 的純安全政策與 Bubble Tea v2 的 TUI state machine。實作需保留既有 Mailer 依賴注入、TUI hostname 支援、單附件選擇與 structured CLI 行為，並確保任何確認狀態都不會自行觸發寄信。

## Goals / Non-Goals

**Goals:**

- 讓 structured CLI 與 Compose TUI 使用相同的單封寄信白名單定義與原因彙整規則。
- 讓 TUI 安全郵件維持一次操作即可送出，白名單外郵件則要求輸入精確 `SEND`。
- 將授權綁定至警告產生時的郵件快照，不允許重用或套用到修改後內容。
- 以純政策測試與 Bubble Tea v2 狀態測試證明確認前 Mailer 呼叫次數為零。

**Non-Goals:**

- 不變更 Burst 的大量寄信 `--allow-domain` 政策。
- 不要求 Compose TUI 的 Host 改用 IPv4；structured CLI 的 IPv4 限制保持不變。
- 不加入 SMTP Auth/TLS、歷史／重送、TUI 多附件或密碼輸入。
- 不重新設計 Compose TUI 的主要版面、欄位順序或既有附件選擇流程。

## Decisions

### 1. 抽出純單封安全評估器，而非讓 TUI 呼叫 StructuredSender

在 `sendmail` 層建立不執行 I/O 的單封安全評估介面，輸入最終 From、To、CC、BCC 與 Port，輸出完整的越界原因或基本格式錯誤。三個安全寄件者、`rd01.softnext.com.tw` 精確網域與 port 25 只在此政策定義一次；`StructuredSender` 改為呼叫此評估器，TUI 也以相同結果決定是否顯示確認。

不直接重用 `StructuredSender.Send`，因為它還負責 CLI 專屬的 IPv4、預設地址與隨機內容，套到手動 TUI 會意外改變既有輸入語意。Burst 也維持獨立政策，避免把單封與大量寄信的授權模型過早抽象成同一套。

### 2. 將安全確認建模為 ComposeModel 的明確狀態

ComposeModel 新增 pending confirmation state，保存待寄送的 `MailCompose` 快照、所有安全原因、確認輸入值與錯誤提示。`Ctrl+S` 流程為：

```text
建立 compose
    │
    ▼
基本地址／port 與安全評估
    ├─ 錯誤 ─────────▶ 顯示錯誤，不呼叫 Mailer
    ├─ 全部安全 ─────▶ 沿用既有非同步寄信
    └─ 有越界原因 ───▶ 進入確認狀態，不呼叫 Mailer
                              │
                       SEND + Enter
                              ▼
                       快照一致才寄送
```

確認畫面使用 Compose TUI 內的 overlay／panel，而不是啟動另一個 tea.Program。這可保留原 model、視窗尺寸與依賴注入，也便於直接測試 `Update` 狀態轉移。

### 3. 使用精確 `SEND`，確認不持久化且不以單鍵替代

只有大小寫完全相符的 `SEND` 加 Enter 才構成授權。Esc 取消；其他文字只顯示確認錯誤。確認成功、取消或寄信結果返回後都清除 pending state，不在 viper、檔案或全域變數保存授權。

相較於 `y` 或再次按 `Ctrl+S`，文字 token 更能避免操作慣性造成誤寄，且不需要新的永久設定。

### 4. 確認期間凍結編輯，送出前仍比較郵件快照

確認狀態攔截一般 Header、Composer、附件與寄信快捷鍵，避免畫面底層被修改。送出前仍以明確欄位與 slice 內容比較目前 compose 與 pending snapshot；若不一致即清除授權並要求重新評估。這是防禦性檢查，不依賴 UI 一定會攔截所有未來新增的 message。

### 5. 明確確認只解除白名單限制，不繞過 Mailer 驗證

地址與 port 格式在安全評估階段即回傳錯誤。附件存在性、可讀性與 MIME 建立仍由既有 fail-closed SMTPMailer 驗證；即使已輸入 `SEND`，附件錯誤仍不得呼叫底層 SMTP。這保持安全政策與郵件組裝責任分離。

### 6. 測試依 Bubble Tea v2 分層進行

- `sendmail` 表格測試覆蓋安全、外部 From/To/CC/BCC、子網域、非 25 port、多原因與格式錯誤，並確認 StructuredSender 結果不變。
- `tui` 直接對 ComposeModel 傳入 `tea.KeyPressMsg`，驗證 safe send、warning、錯誤 token、Esc、精確 `SEND`、重複確認與 snapshot mismatch。
- `View()` 只斷言原因、`SEND` 提示與取消提示等穩定片段，不比較完整 ANSI output。
- 以 recording Mailer 驗證所有未授權路徑呼叫次數為零。

## Risks / Trade-offs

- [白名單外的既有 TUI 操作多一道步驟] → 只對越界原因顯示確認，安全測試網域與 port 25 維持直接寄送。
- [抽出政策時改變 structured CLI 錯誤文字或原因順序] → 先以現有 structured tests 鎖定順序與核心片段，再進行最小抽取。
- [未來新增 TUI message 繞過凍結狀態] → 確認期間集中路由按鍵，送出前再次比較 snapshot。
- [確認後附件才在 Mailer 層失敗] → 保留 SMTP 前完整附件載入與零 SMTP 保證，錯誤回到現有 TUI 結果流程。
- [完整 model view 測試容易受 ANSI 與尺寸影響] → 固定尺寸且只比對必要文字片段。

## Migration Plan

1. 以測試鎖定既有 structured safety 行為，再抽出純安全評估器。
2. 加入 ComposeModel pending confirmation state 與 headless state tests。
3. 加入確認畫面與 View assertions，更新 README 的 TUI 安全操作說明。
4. 執行 `go test ./sendmail ./tui`、`go vet ./...` 與 `make test`。

本變更沒有持久資料遷移。若需回退，可移除 TUI confirmation state 並讓 `Ctrl+S` 回復直接呼叫 Mailer；抽出的純政策仍可供 structured CLI 使用。

## Open Questions

- 無；本 change 採用精確 `SEND`、確認期間凍結編輯、TUI hostname 維持允許的既定範圍。
