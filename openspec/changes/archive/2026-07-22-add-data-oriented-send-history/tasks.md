## 1. History concrete data 與純函式（Red → Green）

- [x] 1.1 在 `sendmail` 新增 history table tests，定義 versioned `HistoryRecord`、resolved request/result JSON shape、唯一 ID、時間與成功／失敗 Mailer attempt 資料。
- [x] 1.2 新增純資料操作測試，覆蓋 records 依時間新到舊排序、預設／指定 limit、ID lookup、summary 不包含 Body/BCC/附件，以及 record-to-structured-options 不攜帶 confirmation/no-history。
- [x] 1.3 新增 validation tests，覆蓋空 ID、重複 ID、缺少必要 request 欄位、不支援 version 與無效 result，並先執行 focused tests 取得 Red 結果。
- [x] 1.4 實作 concrete history structs、schema version、record 建立、validation、summary、recent、lookup 與 replay options 等 ordinary functions；不得新增 Store／Repository interface。
- [x] 1.5 執行 history 純 unit tests，確認資料 round-trip、排序、查找、summary 隱私與 replay 轉換全部通過。

## 2. JSONL 路徑、權限與 fail-closed I/O（Red → Green）

- [x] 2.1 新增預設路徑測試，覆蓋 `$XDG_STATE_HOME/hermes/history.jsonl`、home fallback、缺少 home 錯誤，以及測試可直接傳入 temp path 而不碰真實使用者目錄。
- [x] 2.2 新增 JSONL append/read tests，覆蓋首次建立、兩筆以上順序、空檔、unicode／多行 Body escaping、單一 process concurrent append 與 round-trip。
- [x] 2.3 新增權限與錯誤測試，覆蓋目錄 `0700`、檔案 `0600`、不可寫路徑、截斷 JSON、缺少欄位、不支援 version、重複 ID及錯誤行號，並先執行 focused tests 取得 Red 結果。
- [x] 2.4 實作 history path resolution、單次 JSON-line encode/append、同 process 最小寫入序列化、完整 read/validate 與限制性權限；不得靜默跳過損壞 record。
- [x] 2.5 執行 JSONL focused tests，確認 valid data、concurrency、permissions 與 malformed/version fail-closed 行為全部通過。

## 3. Structured send execution 與紀錄協調（Red → Green）

- [x] 3.1 新增 structured execution tests，鎖定 preflight failure、Mailer success、Mailer error 的 concrete result，並確認既有 `Send(options) error` API 與所有安全／附件行為保持相容。
- [x] 3.2 新增 history coordinator tests，覆蓋預設記錄成功與 SMTP 失敗、preflight 失敗不記錄、send `--no-history` 零 history I/O，以及 record 不含 confirmation 或 secret 欄位。
- [x] 3.3 新增雙結果錯誤測試，確認寄送成功但 append 失敗時明確標示郵件已送出；寄送與 append 皆失敗時保留兩個原因，且 Mailer 永遠只呼叫一次。
- [x] 3.4 實作不破壞 `Send` 的 concrete structured execution result，讓 CLI coordinator 只在 Mailer attempted 後建立／追加 history record；不得以 history-aware Mailer decorator 或新 persistence interface 實作。
- [x] 3.5 在 `hermes send` 加入 `--no-history` 與預設 history path 協調，更新 cmd tests 驗證 flags、help、錯誤傳遞及無意外 TUI/history side effects。
- [x] 3.6 執行 `go test ./sendmail ./cmd`，確認 structured regression、history recording、no-history 與雙結果錯誤全部通過。

## 4. History list／show structured CLI（Red → Green）

- [x] 4.1 新增 command tests，覆蓋 `history list` 預設 10、正整數 `--limit`、新到舊 JSON summaries、空 history，以及 summary 不輸出 Body/BCC/附件。
- [x] 4.2 新增 `history show <id>` tests，覆蓋完整 JSON record、unicode／多行 Body、存在與不存在 ID、錯誤參數數量、損壞 history 及不支援 version。
- [x] 4.3 新增 root routing/help tests，確認 history/list/show 不啟動 TUI、不呼叫 Mailer，且 help 不讀寫 history。
- [x] 4.4 實作可注入 command writer 與明確 history path 的 `history` command group、list/show 子命令及穩定 JSON encoder，並註冊至 root command。
- [x] 4.5 執行 history list/show focused cmd tests，確認輸出、limit、錯誤與零寄信副作用全部通過。

## 5. Safe replay（Red → Green）

- [x] 5.1 新增 replay tests，覆蓋安全 record 成功重建所有 resolved 欄位、預設新增 record、`--no-history` 仍讀取來源但不追加，以及 Mailer 只呼叫一次。
- [x] 5.2 新增安全 regression tests，覆蓋外部 From、To/CC/BCC、非 25 port 未重新確認時拒絕，多原因穩定順序，以及本次 `--confirm-outside-whitelist` 成功授權。
- [x] 5.3 新增 fail-closed replay tests，覆蓋不存在 ID、損壞／不支援 history、無效 server/address/port、缺失附件，並斷言 Mailer、SMTP 與新 record 數量皆為零。
- [x] 5.4 實作 `history replay <id>`、`--confirm-outside-whitelist` 與 `--no-history`；以 record-to-options 純轉換重新呼叫現有 structured pipeline，不得直接組裝 SMTP message 或沿用舊授權。
- [x] 5.5 執行 replay focused tests，確認欄位保真、重新授權、附件 preflight、no-history 與零副作用全部通過。

## 6. Mailpit、文件與完整驗證

- [x] 6.1 新增 Mailpit integration test：structured send 建立 record，再 replay 同一 record，驗證郵件數量、To/CC/BCC envelope、Subject/Body 與多附件 MIME 均一致。
- [x] 6.2 新增 Mailpit fail-closed cases，確認 replay 缺失附件、未重新確認或損壞 record 時 Mailpit 郵件數量不增加；測試 history file 必須使用 temp path。
- [x] 6.3 更新 `README.md`，說明 list/show/replay、JSON output、預設最近 10 筆、history path、Body/BCC/附件路徑敏感性、`0600`、`--no-history` 與 replay 重新授權。
- [x] 6.4 對受影響 Go 檔執行 gofmt，並執行 `go vet ./...`。
- [x] 6.5 執行 `make test`，確認 unit、CLI、history file、Mailpit integration、race detector 與 coverage 全部通過。
