## Why

Burst mode 目前只能從同一地址池隨機產生寄件人與收件人，無法固定其中任一方；同時 CLI 雖讀取 `burst-domain`，卻未提供對應旗標或安全驗證，可能產生無效地址，甚至誤將大量測試信送往正式網域。需要讓地址來源可明確控制，並在任何郵件送出前阻擋未授權網域。

## What Changes

- 新增可選的 `--from` 與 `--to`，有提供時使用並驗證指定地址，未提供時才產生隨機地址。
- 正式提供 `--domain`，且只要 From 或 To 任一方需要隨機產生，就必須明確指定該旗標；不提供隱含預設網域。
- 內建安全網域正向表列，初始僅包含 `rd01.softnext.com.tw`。
- 新增可重複的 `--allow-domain`，讓使用者逐一明確授權本次執行所需的非表列網域。
- 在建立 goroutine 或送出任何郵件前，一次驗證所有固定地址、隨機網域及網域授權；任一項失敗即整批拒絕。
- 保留既有隨機地址、隨機主旨與內容，以及併發發送行為。
- 更新 Burst CLI 說明與使用範例。
- **BREAKING**：只提供 `--quantity`、`--host`、`--port` 的既有 burst 呼叫將不再開始寄信；使用隨機地址時必須另外提供 `--domain`。

## Capabilities

### New Capabilities

- `burst-address-controls`: 定義 burst 固定／隨機地址選擇、網域正向表列、逐網域授權及整批預先驗證行為。

### Modified Capabilities

- `cli-mode-routing`: 更新 burst 子命令的必要參數與安全失敗行為。
- `unified-email-sending`: 更新 Burst mode 從一律隨機地址，改為依明確參數選擇固定或隨機地址並在發送前驗證。

## Impact

- CLI：`cmd/burst_cmd.go` 的旗標、參數解析、錯誤回傳與說明文字。
- 發信邏輯：`sendmail/burst_send.go` 的輸入模型、地址選擇及發送前驗證。
- 共用工具：可能調整 `utils` 的隨機地址產生介面，使網域來源明確且可測試。
- 測試：更新 `cmd` 與 `sendmail` 的單元測試，覆蓋固定／隨機組合、表列與非表列網域、授權及零封信副作用。
- 文件：同步更新 `README.md` 的 Burst mode 參數與安全範例。
- 不新增第三方依賴，不變更 SMTP 通訊協定或 TUI 行為。
