## MODIFIED Requirements

### Requirement: Structured 單封寄送紀錄採版本化資料格式
系統 SHALL 將每次通過 preflight 並實際呼叫 Mailer 的 structured 單封寄送表示為具版本、唯一 ID、時間、resolved request 與 result 的 concrete history record。Request SHALL 保存 server、port、From、To、CC、BCC、Subject、Body、正規化附件路徑，以及 TLS mode、TLS server name、Auth mode 與 username；result SHALL 保存成功狀態與不含 transport secret 的寄送錯誤。

#### Scenario: 成功寄送建立紀錄
- **WHEN** structured send 通過所有驗證且 Mailer 成功完成
- **THEN** 系統追加一筆包含完整可保存 resolved request、transport metadata 與成功結果的 history record

#### Scenario: SMTP 失敗建立紀錄
- **WHEN** structured send 通過 preflight、Mailer 已被呼叫但回傳錯誤
- **THEN** 系統追加一筆失敗 record，保存不含 password 的寄送錯誤且不自動重試

#### Scenario: Preflight 失敗不建立紀錄
- **WHEN** server、地址、port、安全確認、附件或 transport preflight 失敗而未呼叫 Mailer
- **THEN** 系統不建立 history record

#### Scenario: 一次性控制與 secret 不寫入紀錄
- **WHEN** 系統建立 history record
- **THEN** record 不包含 confirmation、no-history、SMTP password、password source 內容、token、已認證狀態或其他 transport secret

### Requirement: History replay 重新執行目前寄送流程
系統 SHALL 提供 `hermes history replay <id>`，將 record request 轉換為 structured send options，再重新執行目前的格式驗證、安全評估、附件與 transport preflight、Mailer 與 history 規則。Replay MUST NOT 直接從 record 呼叫 SMTP，且需要 PLAIN Auth 時 MUST 由本次命令重新取得 password。

#### Scenario: Replay 有效安全紀錄
- **WHEN** 指定 record 仍符合目前規則、附件仍存在、transport 不需 Auth 且 Mailer 成功
- **THEN** 系統以記錄的郵件與 transport metadata 寄送一次，並預設追加一筆新的成功 record

#### Scenario: Replay authenticated record 並重新提供 password
- **WHEN** record 使用 PLAIN Auth，且本次 replay 提供 `--auth-password-stdin` 與非空 password
- **THEN** 系統使用 record 的 Auth mode／username 與本次 password 重新執行完整 transport 流程

#### Scenario: Replay authenticated record 未提供 password
- **WHEN** record 使用 PLAIN Auth，但本次 replay 未提供 `--auth-password-stdin` 或讀到空 password
- **THEN** 系統在 SMTP 連線前拒絕 replay，且不建立新 record

#### Scenario: Replay 停用新紀錄
- **WHEN** 使用者執行有效 replay 並提供 `--no-history`
- **THEN** 系統寄送一次但不追加 replay record

#### Scenario: Replay 附件已不存在
- **WHEN** record 參照的任一附件已不存在或不可讀
- **THEN** 系統在讀取 password、呼叫 Mailer或 SMTP 前拒絕 replay，且不建立新 record

#### Scenario: Replay record 不存在或不可用
- **WHEN** 指定 ID 不存在，或 history data 損壞／版本不支援
- **THEN** 系統回傳錯誤且不讀取 password，Mailer 與 SMTP 呼叫次數皆為零
