# structured-send-plan Specification

## Purpose

定義 structured send 在產生副作用前，如何以完整且具體的 plan 統一郵件內容、SMTP transport、執行結果與寄送歷史，確保正式寄送與測試使用相同資料流並維持既有外部契約。

## Requirements

### Requirement: Structured send 在副作用前建立完整 plan
系統 SHALL 在呼叫寄送 attempt 前建立 concrete structured send plan；plan MUST 同時包含已解析的郵件 message/envelope 與完成 ready validation 的 SMTP transport data。任何 defaults、格式、安全、附件、隨機內容或 transport validation 失敗 SHALL 不得產生可執行 plan 或呼叫 attempt。

#### Scenario: 成功建立完整 plan
- **WHEN** structured options 通過所有 preflight 且缺省內容已補齊
- **THEN** plan 包含最終 From、To、CC、BCC、Subject、Body、附件、server、port、TLS/Auth policy 與本次必要 credential

#### Scenario: Preflight 失敗
- **WHEN** 任一地址、安全授權、附件、transport 組合或 ready credential 驗證失敗
- **THEN** 系統回傳原有具體錯誤，寄送 attempt 呼叫次數為零

### Requirement: Structured execution 永遠使用完整 plan
Structured sender SHALL 透過單一 plan-aware attempt 執行寄送，且 MUST 將同一 plan 的 message 與 transport 一起傳入 I/O 邊界。執行選擇 MUST NOT 依 attempt wrapper 或 mailer 的 concrete type 決定是否套用 transport。

#### Scenario: 正式 SMTP attempt
- **WHEN** structured command 執行有效 plan
- **THEN** SMTPMailer 使用 plan 的 Compose 與 SMTPTransportConfig 完成一次寄送

#### Scenario: 測試記錄 attempt
- **WHEN** 測試注入非 SMTP concrete 的記錄函式並執行 TLS/Auth plan
- **THEN** 記錄函式仍收到完整 TLS/Auth data，且不會退回只含 MailCompose 的路徑

#### Scenario: Attempt 失敗
- **WHEN** plan-aware attempt 回傳錯誤
- **THEN** execution result 保存同一 plan、attempted 狀態與該錯誤，且不自動重試

### Requirement: History 與 attempt 使用同一 plan
系統 SHALL 僅在實際 attempt 後以該 execution 的同一 structured send plan 建立 history record。Record MUST 從 plan 保存 message 與非敏感 transport metadata，MUST NOT 反推、使用預設 transport 取代 plan data，亦 MUST NOT 保存 password。

#### Scenario: 成功 attempt 建立一致 record
- **WHEN** plan-aware attempt 成功且 history 啟用
- **THEN** record 的 message 與 transport metadata 等於 attempt 收到的 plan，並保存成功結果

#### Scenario: 失敗 attempt 建立一致 record
- **WHEN** plan-aware attempt 失敗且 history 啟用
- **THEN** record 保存相同 plan 的非敏感資料與 redacted failure，不執行第二次 attempt

#### Scenario: History 停用
- **WHEN** 本次 structured send 指定 `--no-history`
- **THEN** 系統仍建立並執行相同 plan 一次，但不執行任何 history I/O

#### Scenario: Preflight failure 不建立 record
- **WHEN** plan 建立前失敗而 attempt 未執行
- **THEN** 系統不建立 history record

### Requirement: Structured plan 重構保持既有外部契約
重構後系統 MUST 維持既有 structured CLI flags、預設值、錯誤語意、安全確認、SMTP/TLS/Auth、MIME、history v2 JSON、replay 與 Mailpit observable behavior。

#### Scenario: 既有明文 structured send
- **WHEN** 使用者只提供安全的最小 structured send 參數
- **THEN** 系統維持無 TLS／無 Auth 預設並成功寄送與記錄

#### Scenario: 既有 authenticated structured send
- **WHEN** 使用者提供有效 required STARTTLS、PLAIN Auth 與 stdin password
- **THEN** 系統維持原有 preflight、TLS、Auth、secret handling 與 history metadata 行為

#### Scenario: 既有 history replay
- **WHEN** 使用者 replay 有效 history v2 record 並提供目前需要的確認與 credential
- **THEN** 系統重新建立 plan、只寄送一次，並維持既有新 record 規則
