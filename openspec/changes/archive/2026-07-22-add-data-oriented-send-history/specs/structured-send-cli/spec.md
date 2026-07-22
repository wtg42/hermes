## ADDED Requirements

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
