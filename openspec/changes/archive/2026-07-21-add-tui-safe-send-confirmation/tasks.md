## 1. 共用單封安全政策（Red → Green）

- [x] 1.1 在 `sendmail` 新增安全評估表格測試，覆蓋安全寄件、外部 From、外部 To/CC/BCC、子網域、非 25 port、大小寫與多原因穩定順序。
- [x] 1.2 新增基本格式錯誤測試，覆蓋無效 From/To/CC/BCC、空 To 與超出範圍或非數字 port，確認錯誤不會被視為可授權的越界原因。
- [x] 1.3 補強 StructuredSender regression tests，鎖定三個安全寄件者、精確安全網域、確認旗標與既有原因內容，並執行 focused tests 取得 Red 結果。
- [x] 1.4 實作具名的純單封安全評估輸入／結果，集中安全常數、地址／port 驗證、精確網域比較與原因彙整，不執行任何 I/O。
- [x] 1.5 將 StructuredSender 改為使用共用評估器，保留 CLI 專屬 IPv4、預設值、隨機內容、附件 preflight 與 `--confirm-outside-whitelist` 行為。
- [x] 1.6 執行 `go test ./sendmail`，確認共用政策與 structured regression tests 全部通過。

## 2. ComposeModel 安全確認狀態（Red → Green）

- [x] 2.1 在 `tui` 新增 recording Mailer 與固定 ComposeModel 測試 helper，為安全／外部郵件填入明確 From、To、CC、BCC、Host 與 Port。
- [x] 2.2 新增 `Ctrl+S` 分流測試：安全郵件直接建立一個寄信 command；格式錯誤顯示錯誤且無 command；白名單外郵件進入 pending confirmation 且 Mailer 呼叫次數為零。
- [x] 2.3 新增確認輸入測試，覆蓋空值、`send`、其他文字、精確 `SEND`、Enter、Esc 與確認不可重用，並先執行 focused tests 取得 Red 結果。
- [x] 2.4 在 ComposeModel 加入 pending confirmation 資料、確認輸入與錯誤狀態，將 compose 建立抽成可重用 helper，並在 `Ctrl+S` 前呼叫共用安全評估器。
- [x] 2.5 實作確認狀態的優先按鍵路由：精確 `SEND` 才建立既有非同步寄信 command，Esc 取消，其餘輸入不得觸發 Header、Composer、Filepicker、模板或第二次寄信。
- [x] 2.6 執行安全分流與確認 focused tests，確認 pending、取消、成功與重複操作的 state transition 全部通過。

## 3. 快照綁定、畫面與 fail-closed（Red → Green）

- [x] 3.1 新增快照測試，涵蓋 From、To/CC/BCC、Subject、Body、Host、Port、單附件與多附件 slice 任一變更時舊授權失效且不呼叫 Mailer。
- [x] 3.2 新增快捷鍵隔離測試，確認 pending state 中的 `Ctrl+S`、`Ctrl+A`、`Ctrl+H`、`Ctrl+T`、`Ctrl+E` 與一般文字不修改底層 compose 或開啟 overlay。
- [x] 3.3 實作明確郵件快照比較與確認期間凍結；snapshot mismatch 時清除舊授權、顯示需重新評估的錯誤且不建立寄信 command。
- [x] 3.4 新增 `View()` 片段測試並實作確認 overlay／panel，顯示全部越界原因、`Type SEND to confirm`、輸入錯誤與 Esc 取消提示。
- [x] 3.5 新增已確認但附件不存在的回歸測試，驗證既有 SMTPMailer 回傳附件路徑錯誤且底層 SMTP 呼叫次數為零。
- [x] 3.6 執行 `go test ./tui ./sendmail`，確認安全確認、快照、畫面與附件 fail-closed 行為通過。

## 4. 文件與完整驗證

- [x] 4.1 更新 `README.md` 的 TUI 說明，列出共用白名單、何時需要輸入精確 `SEND`、Esc 取消、確認不持久化，以及 TUI hostname 仍允許。
- [x] 4.2 對受影響 Go 檔執行 gofmt，並執行 `go vet ./...`。
- [x] 4.3 執行 `make test`，確認 unit、Bubble Tea state/View、Mailpit integration、race detector 與 coverage 全部通過。
