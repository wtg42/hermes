## Why

Hermes 目前以未提供 Auth 的 `net/smtp` 路徑寄信，無法讓使用者明確要求 TLS 或驗證認證結果；TUI 甚至在只設定 host 時就顯示 `TLS active`，與實際連線狀態不一致。下一步需要把傳輸安全變成可驗證、失敗即停止的明確行為，同時避免憑證進入寄信歷史或診斷輸出。

## What Changes

- 新增具體、資料導向的 SMTP transport 設定，將連線位址、TLS policy 與 Auth 資料和郵件內容分離。
- 為 structured CLI 新增明確的 TLS 與 SMTP Auth 參數；初始範圍支援無 TLS、強制 STARTTLS，以及無認證、PLAIN 認證。
- 強制 STARTTLS 時要求可驗證的 TLS server name；TCP 連線目標仍必須是使用者明確指定的 IPv4。
- TLS 升級、憑證驗證或 SMTP 認證任一步驟失敗時，在傳送 envelope/message 前停止。
- 密碼只供當次傳輸使用，不得保存到 history、輸出到錯誤／狀態訊息或被 replay 重用。
- 修正 Compose TUI 的連線狀態文字，不再於尚未建立且驗證 TLS 連線時宣稱 `TLS active`。
- 不在本次加入 implicit TLS、OAuth、通用 transport interface 或 TUI Auth/TLS 輸入介面。

## Capabilities

### New Capabilities

- `smtp-transport-security`: 定義 SMTP 連線、STARTTLS、憑證驗證、PLAIN Auth、失敗關閉及敏感資料處理行為。

### Modified Capabilities

- `structured-send-cli`: 擴充 structured CLI 的 transport flags、參數組合與錯誤行為。
- `direct-send-safety`: 在維持明確 IPv4 連線目標下，加入 TLS server name 與認證參數的 preflight 驗證。
- `send-history`: 定義 transport metadata 的保存範圍，並禁止保存或 replay 認證秘密。
- `compose-tui`: 讓畫面只呈現可由目前狀態證明的連線／TLS 資訊。

## Impact

- 影響 `cmd/send_cmd.go` 的 CLI flags、解析與 preflight。
- 影響 `sendmail/` 的 SMTP 連線流程、資料模型、history schema 與 replay 行為。
- 影響 `tui/compose.go` 的連線狀態顯示文字。
- 需要新增不連外的 SMTP/TLS/Auth 測試伺服器或等價測試 fixture；既有 Mailpit MIME／投遞測試維持不變。
- 不新增第三方 SMTP client 依賴，除非實作階段證明標準函式庫無法在合理複雜度內符合規格。
