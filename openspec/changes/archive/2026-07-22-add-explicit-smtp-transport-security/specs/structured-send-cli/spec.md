## MODIFIED Requirements

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

## ADDED Requirements

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
