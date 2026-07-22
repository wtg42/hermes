# structured-send-cli Specification

## Purpose
定義不啟動 TUI 的單封 structured CLI 寄信介面、參數集合、安全預設、測試內容與 MIME 行為。

## Requirements

### Requirement: Structured send command
系統 SHALL 提供 `hermes send` 子命令，讓使用者只透過 flags 建立並發送一封郵件，不啟動任何 TUI。命令 SHALL 支援 `--server`、`--port`、`--from`、`--to`、`--cc`、`--bcc`、`--subject`、`--body`、`--attach`、`--confirm-outside-whitelist`、`--tls-mode`、`--tls-server-name`、`--auth-mode`、`--auth-username` 與 `--auth-password-stdin`。

#### Scenario: 使用最小安全參數寄信
- **WHEN** 使用者只提供有效的 `--server` IPv4
- **THEN** 系統套用安全 From、不同的安全 To、port 25、隨機 Subject/Body、無 TLS 與無 Auth，並發送一封郵件

#### Scenario: 使用完整 structured 參數寄信
- **WHEN** 使用者提供 server、port、From、多個 To/CC/BCC、Subject、Body、已存在附件及有效 transport 參數，且所有安全邊界皆通過或已明確確認
- **THEN** 系統使用解析後的完整郵件與 transport 參數發送一封郵件，不啟動 TUI

#### Scenario: Structured send 失敗
- **WHEN** structured 郵件參數、transport 參數或寄信流程回傳錯誤
- **THEN** `hermes send` 回傳具體且不含 password 的錯誤，並以非零狀態結束

### Requirement: Collection flags 支援多值
`--to`、`--cc`、`--bcc` 與 `--attach` SHALL 接受重複 flags 或逗號分隔值，並保持使用者輸入順序傳入發信流程。

#### Scenario: 重複指定收件人與附件
- **WHEN** 使用者重複提供 To、CC、BCC 或 Attach flags
- **THEN** 系統將所有值納入同一封郵件，BCC 只出現在 SMTP envelope 而不出現在可見 header

#### Scenario: 逗號分隔 collection 值
- **WHEN** 使用者在單一 collection flag 中提供多個逗號分隔值
- **THEN** 系統將每個非空值解析為獨立收件人或附件

### Requirement: Structured send 安全預設
系統 SHALL 在 From、To、port、Subject 或 Body 未指定時套用已定義的安全／測試預設，且預設 From 與 To MUST 不同。

#### Scenario: From 與 To 都未指定
- **WHEN** 使用者未提供 `--from` 與 `--to`
- **THEN** From 使用 `weitingshih@rd01.softnext.com.tw`，To 使用另一個安全地址，兩者不相同

#### Scenario: 指定 From 但未指定 To
- **WHEN** 使用者提供安全寄件者但未提供 To
- **THEN** 系統從安全地址中選擇與最終 From 不同的預設 To

#### Scenario: 未指定 port
- **WHEN** 使用者未提供 `--port`
- **THEN** 系統使用 port 25

### Requirement: 隨機測試內容
系統 SHALL 僅為未提供的 Subject 或 Body 產生測試內容。隨機 Subject SHALL 包含 emoji 與中英文片語；隨機 Body SHALL 包含中英文內容、台北時間與 8 字元十六進位 Trace-ID。

#### Scenario: Subject 與 Body 都未指定
- **WHEN** 使用者未提供 Subject 與 Body
- **THEN** 系統產生隨機 Subject 與 Body，Body 包含 `送出時間 Sent at: YYYY-MM-DD HH:mm:ss +08:00` 及 `Trace-ID: xxxxxxxx`

#### Scenario: 只提供 Subject
- **WHEN** 使用者提供 Subject 但未提供 Body
- **THEN** 系統保留原 Subject，只產生隨機 Body

#### Scenario: 只提供 Body
- **WHEN** 使用者提供 Body 但未提供 Subject
- **THEN** 系統保留 Body 的精確內容且不附加時間或 Trace-ID，只產生隨機 Subject

#### Scenario: Subject 與 Body 都提供
- **WHEN** 使用者同時提供 Subject 與 Body
- **THEN** 系統不修改兩者的內容

### Requirement: Structured send 沿用 Hermes MIME 能力
系統 SHALL 使用共用 Mailer 與 multipart MIME 組裝發送 structured message，保留中文主旨編碼、plain text、HTML、To/CC header 及 BCC envelope 行為。

#### Scenario: Structured 中文郵件
- **WHEN** structured send 的 Subject 或 Body 包含中文與 emoji
- **THEN** 郵件以既有 UTF-8／base64 MIME 規則組裝，Mailpit 可正確解析內容

### Requirement: Structured send history control
`hermes send` SHALL 在 structured message 通過 preflight 並實際進入 Mailer 後，預設保存 resolved request 與寄送結果；命令 MUST 提供 `--no-history` 以完全停用本次 history I/O。

#### Scenario: Structured send 預設記錄
- **WHEN** 使用者執行未指定 `--no-history` 的 `hermes send` 且 Mailer 被呼叫
- **THEN** 系統在 Mailer 完成後保存一筆成功或失敗 history record

#### Scenario: Structured send 明確不記錄
- **WHEN** 使用者執行帶有 `--no-history` 的 `hermes send`
- **THEN** 系統沿用相同預設值、驗證、安全與 SMTP 行為，但不執行任何 history I/O

#### Scenario: Structured preflight 失敗不記錄
- **WHEN** structured send 在呼叫 Mailer 前因任何驗證失敗而結束
- **THEN** 系統不建立 history record，並維持既有具體錯誤與非零狀態

#### Scenario: History 寫入錯誤不重送
- **WHEN** Mailer 已完成但 history append 失敗
- **THEN** 系統回報可區分寄送與 history 狀態的錯誤，且不再次呼叫 Mailer

### Requirement: Structured send 安全讀取 SMTP password
`hermes send` SHALL 僅在指定 `--auth-password-stdin` 時從標準輸入讀取一行 password，並 SHALL 移除該行的換行結尾。命令 MUST NOT 提供接受 password value 的 CLI flag，且 MUST 在其他靜態 preflight 全部通過後才讀取 stdin。

#### Scenario: 從 stdin 提供 password
- **WHEN** 使用者選擇 PLAIN Auth、提供 username 與 `--auth-password-stdin`，且其他 preflight 通過
- **THEN** 系統讀取一行非空 password 供本次 Auth 使用，不將其輸出或持久化

#### Scenario: PLAIN Auth 未要求讀取 stdin
- **WHEN** 使用者選擇 PLAIN Auth 但未提供 `--auth-password-stdin`
- **THEN** 系統在連線前回傳缺少 password source 的錯誤

#### Scenario: stdin password 為空
- **WHEN** `--auth-password-stdin` 讀到空值或只有換行
- **THEN** 系統在連線前拒絕寄送，且錯誤不回顯輸入內容

#### Scenario: 其他 preflight 先失敗
- **WHEN** Email、附件、server、port、安全授權或 transport 組合無效
- **THEN** 系統不讀取 stdin、不呼叫 Mailer且不建立 SMTP 連線
