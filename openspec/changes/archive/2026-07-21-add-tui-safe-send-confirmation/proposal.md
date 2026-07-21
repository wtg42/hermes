## Why

`hermes send` 與 Burst 已具備寄件安全邊界，但一般 Compose TUI 仍會在按下 `Ctrl+S` 後直接呼叫 Mailer，可能把測試郵件寄往正式或外部信箱。由於使用者已將避免誤寄列為優先需求，TUI 必須在 SMTP 前套用與 structured 單封寄信一致的白名單評估與明確授權。

## What Changes

- 將 structured 單封寄信既有的安全寄件者、安全收件網域與安全 port 評估抽成可由 TUI 共用的單封安全政策。
- Compose TUI 在寄信前彙整所有白名單外原因；安全郵件維持直接送出，白名單外郵件則進入確認狀態且不呼叫 Mailer。
- 白名單外寄信要求使用者輸入精確的 `SEND` 後才允許送出，Esc 可取消並返回原撰寫內容。
- 確認只綁定觸發警告時的郵件快照；若內容或 Header 已變更，舊確認不得授權新的郵件。
- 保留 TUI hostname 支援與既有單附件介面，不改變 Burst 的大量寄信安全政策。

## Capabilities

### New Capabilities

- 無。

### Modified Capabilities

- `direct-send-safety`: 新增一般 Compose TUI 共用白名單評估、明確確認與快照綁定的安全需求。
- `compose-tui`: 修改 `Ctrl+S` 發送流程，加入安全警告／確認狀態、取消行為與零副作用保證。

## Impact

- 主要影響 `sendmail/structured_send.go` 的安全政策邊界，以及 `tui/compose.go` 的狀態、輸入處理與畫面渲染。
- 需新增純函式安全政策測試與 Bubble Tea v2 `Update`／`View` 狀態測試，並確認既有 structured CLI 行為不變。
- 不新增依賴，不加入 SMTP Auth/TLS、歷史重送、TUI 多附件或 Burst 行為變更。
