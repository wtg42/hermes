## Context

Hermes 的 `hermes send` 已將 flags 解析、預設值、格式與安全評估、附件 preflight、隨機內容及 Mailer 呼叫集中在 `StructuredSender`。目前 `Send` 只回傳 error，成功或 SMTP 失敗後沒有可供查詢與 replay 的 resolved message/result 資料。專案執行於 Linux 開發環境，並維持 Go 標準函式庫、TDD、零產品 SMTP 測試與資料導向優先的設計原則。

History 會保存收件者、BCC、Body 與附件路徑，屬敏感本機資料；replay 又具有寄信副作用，因此資料格式、檔案權限、損壞容錯、安全重新評估與寄送／寫檔結果的區分必須在實作前明確定義。

## Goals / Non-Goals

**Goals:**

- 以 concrete records、slice 與純轉換函式表示 history，不為單一 JSONL backend 建立 Store／Repository interface。
- 只記錄已通過 preflight 並實際呼叫 Mailer 的 structured 單封寄送，包括 SMTP 成功與失敗結果。
- 提供 agent-friendly 的 list、show 與 replay CLI，所有讀取結果具有穩定 JSON 表示。
- Replay 重建原 resolved message，但重新執行現在的格式、安全與附件驗證，且不沿用舊授權。
- 使用版本化 JSONL、限制性權限與 fail-closed 解析，避免不完整或不支援資料被寄送。
- 保持既有 `Mailer` 介面與 TUI／Burst／EML 行為不變。

**Non-Goals:**

- 不建立可插拔 storage backend、資料庫、遠端同步或通用 persistence framework。
- 不為 TUI、Burst 或 EML 增加 history。
- 不保存附件內容、SMTP credentials、token、Auth/TLS secret 或白名單確認狀態。
- 不新增 history 自動清理、刪除、編輯、匯入、匯出或 `last` 快捷重送。
- 不保證讀取未來版本或人工修改後仍可 replay。

## Decisions

### 1. History 採版本化 concrete data，不建立 storage interface

新增具名 concrete structs，例如 `HistoryRecord`、`RecordedSendRequest` 與 `RecordedSendResult`。資料處理以 ordinary functions 完成：record 建立、JSON encode/decode、recent slicing、ID lookup 與 replay options 轉換。單一檔案 backend 直接以路徑參數呼叫 `AppendHistory`／`ReadHistory` 類型的函式。

這符合專案資料導向原則，讓測試直接傳入 record slice 或 temp file。替代方案是先建立 `HistoryStore`、`HistoryRepository` 及 fake implementation，但目前沒有第二個 backend，也沒有需要跨 package 動態替換的真實行為，因此不採用。若未來出現 SQLite 或遠端 backend，再由實際使用端抽出最小介面。

### 2. JSONL record 保存 resolved request 與一次 Mailer attempt 結果

每行是一筆獨立 JSON object，包含 schema version、不可預測 ID、建立時間、resolved server/port、From、To/CC/BCC、Subject、Body、正規化附件路徑，以及 success/error result。欄位使用明確 struct 與 JSON tags，不直接序列化可能持續擴張的 `StructuredSendOptions` 或 `MailCompose`。

只在所有 preflight 通過、Mailer 確實被呼叫後建立 record；格式、安全或附件 preflight 失敗不記錄。SMTP／Mailer error 會保存可顯示的 error string。確認旗標與 `--no-history` 是一次性控制資料，不寫入 record；未來 Auth/TLS credentials 也明確排除。

JSONL 可用單次 append 保留既有紀錄，讀取後以 slice 做排序、limit 與查找。替代方案是單一 JSON array，但每次寄送都需重寫完整檔案，且中途中斷影響範圍較大，因此不採用。

### 3. 預設 state path 與權限集中為具名路徑規則

預設檔案位於 `$XDG_STATE_HOME/hermes/history.jsonl`；未設定時使用 `~/.local/state/hermes/history.jsonl`。目錄建立為 `0700`，檔案建立及每次使用時確保為 `0600`。路徑解析與 I/O 接受明確 path 參數，使 unit tests 使用 `t.TempDir()`，不讀寫真實使用者 home。

寫入先將整筆 record marshal 成單一 JSON line，再以 append mode 完成一次 Write；同一 process 的 append 由最小 critical section 序列化。讀取遇到截斷、空 ID、重複 ID、不支援 version 或無效 JSON 時回傳包含行號的錯誤，不回傳可 replay 的部分結果。跨 process 強一致 locking 與自動修復不在本次範圍，文件提醒勿人工並行修改。

### 4. Structured send 回傳 concrete execution data，再由 CLI 協調 history

在不破壞既有 `Send(options) error` 呼叫端的前提下，新增可回傳 concrete execution result 的路徑；結果明確區分 preflight 未通過、Mailer 已嘗試、Mailer error 與 resolved message。`Send` 保持相容並委派此流程，`cmd` 的 send/replay coordinator 則在 `MailerAttempted` 時建立 record。

History 不包裝成另一層 Mailer decorator，因為 decorator 難以辨識 preflight、resolved defaults 與一次性安全授權，也容易讓 TUI／其他 Mailer 使用者意外開始記錄。CLI coordinator 直接處理 concrete execution data，比新增 history-aware Mailer interface 更可預測。

若寄送成功但 history 寫入失敗，命令 MUST 回報「郵件已送出但歷史寫入失敗」的明確錯誤，避免使用者誤以為 SMTP 未執行；若寄送與 history 都失敗，錯誤同時保留兩個原因。History 失敗不得觸發自動重寄。

### 5. List/show 使用穩定 JSON，limit 僅影響輸出

`history list` 預設將依時間新到舊的最近 10 筆 summary 輸出為 JSON array，`--limit` 必須是正整數；summary 不含 Body，但包含 ID、時間、server、From、To、Subject 與結果。`history show <id>` 輸出一筆完整 JSON record，包括 CC/BCC、Body、附件路徑與 error。

輸出使用 command writer，不直接 `fmt.Print` 到全域 stdout，方便 cmd tests。History 檔不因 list limit 被截斷。人類表格、搜尋及刪除可在後續 change 增加，本次以結構穩定與 agent 可解析為優先。

### 6. Replay 是 record-to-options 純轉換加既有寄送 pipeline

`history replay <id>` 先完整讀取並驗證 history，再將 `RecordedSendRequest` 純轉換為 `StructuredSendOptions`。轉換保留 server、port、地址、Subject、Body 與附件路徑，但 confirmation 永遠為 false；只有本次命令提供 `--confirm-outside-whitelist` 才能授權白名單外 replay。

Replay 呼叫與 `hermes send` 相同的 structured resolution、安全評估、附件 preflight 與 Mailer。附件已刪除、地址規則改變、record server 無效或目前安全政策要求確認時，皆在 SMTP 前拒絕。成功進入 Mailer 的 replay 預設追加新 record，`--no-history` 可停用。

直接從 record 組裝 `MailCompose` 並呼叫 SMTP 的替代方案會繞過現在的 policy，因此不採用。

### 7. 測試分層且不以 interface mock history

- 純 unit tests：record 建立、JSON round-trip、recent/lookup、record-to-options、版本與 malformed data。
- 檔案 tests：temp directory、append 順序、權限、空檔、重複 ID、單一 process concurrent append。
- cmd tests：注入 concrete execution function與 temp history path，驗證 flags、JSON output、錯誤及呼叫次數；不建立 history repository mock。
- Mailpit integration：發一封 structured message、讀取 record、replay 並驗證第二封 MIME；另驗證缺失附件與未重新確認時 Mailpit 數量不增加。

不連線產品 SMTP，不將本機 Linux 權限測試視為 FreeBSD 產品驗收；Hermes 本身目前是本機工具，CI 仍執行相同 Go unit/integration suite。

## Risks / Trade-offs

- [Risk] History 保存 BCC、Body 與附件路徑可能洩漏敏感資訊 → 使用 `0700`／`0600`、提供 `--no-history`、README 明確警告，永不保存 credentials。
- [Risk] 寄送成功但 history 寫入失敗造成使用者重試並寄出重複郵件 → 錯誤明確標記 SMTP 已完成，禁止自動 retry。
- [Risk] JSONL 某行損壞使查詢暫時不可用 → 回報行號並 fail-closed；本次不靜默跳過或 replay 部分資料。
- [Risk] 跨 process 同時 append 可能受檔案系統語意影響 → 每筆使用單次 append Write，同 process 序列化；跨 process locking 留待有真實需求時處理。
- [Risk] History 檔無上限成長 → `list --limit` 控制讀取輸出但不清除檔案；retention/delete 另開 change。
- [Risk] Concrete file functions 未來改用資料庫時需要重構 → 目前只有一個 backend，保持 record 與純轉換獨立可降低後續替換成本。

## Migration Plan

1. 先以 table tests 定義 record schema、純轉換、JSONL 讀寫、權限與 malformed/version 行為。
2. 擴充 structured send 產生 concrete execution data，保留既有 `Send` API 與所有安全測試。
3. 實作 send history coordinator、`--no-history` 與 history list/show CLI。
4. 實作 replay 轉換與重新安全評估，再補 Mailpit integration tests。
5. 更新 README，執行 gofmt、go vet 與 `make test`。

Rollback 可移除 history commands/coordinator 並保留既有 structured send；history JSONL 是使用者本機資料，不自動刪除。沒有 DB migration。

## Open Questions

無。若未來需要 retention、跨 process locking、TUI history 或多 backend，分別建立新 change，以實際需求決定抽象。
