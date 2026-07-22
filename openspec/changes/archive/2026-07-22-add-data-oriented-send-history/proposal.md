## Why

Hermes 已具備安全的 structured 單封寄信流程，但仍無法查詢最近寄送內容或以相同資料重新測試；這使操作員與 agent 難以重現 SMTP 問題。新增本機寄送歷史可補齊 `$sendmail-tool` 的核心能力，並以資料導向設計建立明確、可測試且不過度抽象的紀錄與 replay 流程。

## What Changes

- 新增 `hermes history list`、`hermes history show <id>` 與 `hermes history replay <id>` structured CLI，不啟動 TUI。
- `hermes send` 預設記錄已通過 preflight 並實際進入 Mailer 的單封寄送資料與結果；新增 `--no-history` 供敏感或 agent-driven 寄送明確停用紀錄。
- 使用具版本欄位的 concrete history record 與 JSON Lines 本機檔案保存 resolved message、寄送時間及成功／失敗結果；不建立通用 Store／Repository interface。
- 歷史檔與新建目錄採限制性權限；不得保存 SMTP credential、token 或未來 transport secret，文件清楚說明 Body、收件者與附件路徑會被保存。
- `list` 預設顯示最近 10 筆摘要並允許限制筆數；`show` 顯示指定紀錄的完整可保存內容與結果。
- `replay` 從紀錄重建 structured send options，重新執行目前的地址、server、port、白名單及附件 preflight；舊紀錄不得成為安全授權，白名單外郵件必須再次明確確認。
- Replay 預設建立新的 history record，並支援 `--no-history`；不存在、損壞或不支援版本的紀錄必須在寄信前拒絕。
- 本次不新增 TUI history 畫面、burst／EML history、SMTP Auth/TLS、credential 保存、附件內容複製、自動清理或多種 storage backend。

## Capabilities

### New Capabilities

- `send-history`: 定義資料導向的本機單封寄送紀錄格式、list／show／replay 行為、隱私權限、版本容錯與 replay 安全邊界。

### Modified Capabilities

- `cli-mode-routing`: 新增不啟動 TUI 的 `history` command group 與 list／show／replay 子命令路由。
- `structured-send-cli`: 新增 structured send 的預設紀錄與 `--no-history` opt-out 行為，並區分寄送結果與 history 寫入結果。
- `direct-send-safety`: 明定 history replay 必須重新套用目前的單封安全評估，且不得沿用舊確認授權。

## Impact

- CLI：新增 history command group，擴充 `send` 與 replay flags、help 及錯誤輸出。
- 寄送邏輯：將 resolved structured message 與寄送結果表示為可記錄的 concrete data，保持既有 Mailer 邊界與零 SMTP 副作用驗證。
- 本機狀態：於使用者 state 目錄建立 JSONL history 檔，不在 repo 或 viper compose state 保存紀錄。
- 測試：新增純資料轉換、JSONL 權限／容錯、CLI list/show/replay、no-history、安全重新評估及 Mailpit replay 整合測試。
- 文件：說明 history 預設、敏感內容、檔案位置、停用方式與 replay 風險。
- 依賴：不新增第三方 runtime dependency。
