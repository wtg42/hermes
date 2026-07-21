## ADDED Requirements

### Requirement: Structured send command
系統 SHALL 提供 `hermes send` 子命令，讓使用者只透過 flags 建立並發送一封郵件，不啟動任何 TUI。命令 SHALL 支援 `--server`、`--port`、`--from`、`--to`、`--cc`、`--bcc`、`--subject`、`--body`、`--attach` 與 `--confirm-outside-whitelist`。

#### Scenario: 使用最小安全參數寄信
- **WHEN** 使用者只提供有效的 `--server` IPv4
- **THEN** 系統套用安全 From、不同的安全 To、port 25 及隨機 Subject/Body，並發送一封郵件

#### Scenario: 使用完整 structured 參數寄信
- **WHEN** 使用者提供 server、port、From、多個 To/CC/BCC、Subject、Body 與已存在附件，且所有安全邊界皆通過或已明確確認
- **THEN** 系統使用解析後的完整參數發送一封郵件，不啟動 TUI

#### Scenario: Structured send 失敗
- **WHEN** structured 參數或寄信流程回傳錯誤
- **THEN** `hermes send` 回傳具體錯誤並以非零狀態結束

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
