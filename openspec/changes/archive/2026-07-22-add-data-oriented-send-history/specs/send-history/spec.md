## ADDED Requirements

### Requirement: Structured 單封寄送紀錄採版本化資料格式
系統 SHALL 將每次通過 preflight 並實際呼叫 Mailer 的 structured 單封寄送表示為具版本、唯一 ID、時間、resolved request 與 result 的 concrete history record。Request SHALL 保存 server、port、From、To、CC、BCC、Subject、Body 與正規化附件路徑；result SHALL 保存成功狀態與寄送錯誤。

#### Scenario: 成功寄送建立紀錄
- **WHEN** structured send 通過所有驗證且 Mailer 成功完成
- **THEN** 系統追加一筆包含完整 resolved request 與成功結果的 history record

#### Scenario: SMTP 失敗建立紀錄
- **WHEN** structured send 通過 preflight、Mailer 已被呼叫但回傳錯誤
- **THEN** 系統追加一筆失敗 record，保存寄送錯誤且不自動重試

#### Scenario: Preflight 失敗不建立紀錄
- **WHEN** server、地址、port、安全確認或附件 preflight 失敗而未呼叫 Mailer
- **THEN** 系統不建立 history record

#### Scenario: 一次性控制與 secret 不寫入紀錄
- **WHEN** 系統建立 history record
- **THEN** record 不包含 confirmation、no-history、SMTP credential、token 或 transport secret

### Requirement: History 使用受限權限的本機 JSONL
系統 SHALL 將 history records 以每行一筆 JSON object 保存於使用者 state 目錄。預設路徑 MUST 為 `$XDG_STATE_HOME/hermes/history.jsonl`，未設定該環境值時 MUST 使用 `~/.local/state/hermes/history.jsonl`；Hermes 建立的目錄與檔案權限 MUST 分別限制為 `0700` 與 `0600`。`--no-history` MUST 停用本次新增 record 的寫入，但 replay 仍可讀取來源 record。

#### Scenario: 首次寫入建立安全路徑
- **WHEN** 預設 history 目錄與檔案尚不存在且需保存 record
- **THEN** 系統建立目錄與檔案、套用限制性權限並追加有效 JSON line

#### Scenario: 多筆紀錄保持獨立順序
- **WHEN** 系統依序追加多筆 history records
- **THEN** 每筆 record 佔一行且讀取結果保持 append 順序，不重寫或遺失既有有效紀錄

#### Scenario: Send history 明確停用
- **WHEN** 本次 send 指定 `--no-history`
- **THEN** 系統不建立目錄、不讀寫 history file，且寄送行為不因 history 停用而改變

#### Scenario: Replay 新紀錄明確停用
- **WHEN** 本次 replay 指定 `--no-history`
- **THEN** 系統可讀取來源 record 完成 replay，但不建立或追加新的 history record

### Requirement: History 讀取必須驗證完整資料
系統 MUST 在 list、show 或 replay 前驗證每筆 record 的 JSON、schema version、必要欄位與唯一 ID。任何截斷、重複 ID、無效資料或不支援版本 SHALL 回傳含行號或 record 識別資訊的錯誤，且不得提供可 replay 的部分結果。

#### Scenario: 讀取有效 JSONL
- **WHEN** history file 僅包含目前支援版本的完整 records
- **THEN** 系統解碼全部資料並提供後續排序、查找與 replay 使用

#### Scenario: History 包含損壞行
- **WHEN** 任一 JSON line 截斷、無效或缺少必要欄位
- **THEN** 系統回傳指出問題行的錯誤，不靜默略過該行

#### Scenario: History 包含不支援版本
- **WHEN** record schema version 不是目前支援版本
- **THEN** 系統拒絕該次讀取與 replay，且不修改 history file

#### Scenario: History 包含重複 ID
- **WHEN** 兩筆 records 使用相同 ID
- **THEN** 系統回傳資料完整性錯誤，不選擇其中任一筆進行 replay

### Requirement: History list 與 show 提供穩定 JSON 輸出
系統 SHALL 提供 `hermes history list` 與 `hermes history show <id>`。List SHALL 依時間由新到舊輸出 JSON summary array、預設限制為最近 10 筆且接受正整數 `--limit`；show SHALL 輸出指定 ID 的完整 JSON record。

#### Scenario: 列出預設最近十筆
- **WHEN** history 超過 10 筆且使用者執行 `hermes history list`
- **THEN** 系統輸出最新 10 筆 summary，且不在 summary 顯示 Body、BCC 或附件路徑

#### Scenario: 指定 list limit
- **WHEN** 使用者執行 `hermes history list --limit 3`
- **THEN** 系統輸出最新 3 筆 summary，且不刪除檔案中的其他紀錄

#### Scenario: 顯示完整紀錄
- **WHEN** 使用者執行 `hermes history show <id>` 且 ID 存在
- **THEN** 系統輸出該 record 的完整可保存 request 與 result JSON

#### Scenario: 查詢不存在 ID
- **WHEN** 使用者執行 `hermes history show <id>` 且 ID 不存在
- **THEN** 系統回傳包含該 ID 的錯誤並以非零狀態結束

### Requirement: History replay 重新執行目前寄送流程
系統 SHALL 提供 `hermes history replay <id>`，將 record request 轉換為 structured send options，再重新執行目前的格式驗證、安全評估、附件 preflight、Mailer 與 history 規則。Replay MUST NOT 直接從 record 呼叫 SMTP。

#### Scenario: Replay 有效安全紀錄
- **WHEN** 指定 record 仍符合目前規則、附件仍存在且 Mailer 成功
- **THEN** 系統以記錄的 server、port、地址、Subject、Body 與附件寄送一次，並預設追加一筆新的成功 record

#### Scenario: Replay 停用新紀錄
- **WHEN** 使用者執行有效 replay 並提供 `--no-history`
- **THEN** 系統寄送一次但不追加 replay record

#### Scenario: Replay 附件已不存在
- **WHEN** record 參照的任一附件已不存在或不可讀
- **THEN** 系統在呼叫 Mailer／SMTP 前拒絕 replay，且不建立新 record

#### Scenario: Replay record 不存在或不可用
- **WHEN** 指定 ID 不存在，或 history data 損壞／版本不支援
- **THEN** 系統回傳錯誤且 Mailer 與 SMTP 呼叫次數皆為零

### Requirement: 寄送與 history 寫入結果必須可區分
系統 MUST 分別追蹤 Mailer attempt 與 history append 結果，不得因 history 寫入失敗自動重寄或將已成功的 SMTP 誤報為未執行。

#### Scenario: 寄送成功但 history 寫入失敗
- **WHEN** Mailer 成功但 history record 無法寫入
- **THEN** 命令回傳明確指出郵件已送出但 history 寫入失敗的錯誤，且不再次呼叫 Mailer

#### Scenario: 寄送與 history 寫入皆失敗
- **WHEN** Mailer 回傳錯誤且失敗 record 也無法寫入
- **THEN** 命令回傳可區分兩個失敗原因的錯誤，且 Mailer 只被呼叫一次
