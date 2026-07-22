## ADDED Requirements

### Requirement: History replay 必須取得新的安全授權
History record MUST NOT 保存或重用先前的白名單確認。每次 replay SHALL 以 record 的 server、port、From、To、CC、BCC 與附件重新執行目前的 structured 單封驗證及安全政策。

#### Scenario: Replay 安全範圍內紀錄
- **WHEN** record 的 server、地址、port 與附件仍符合目前所有驗證及安全邊界
- **THEN** replay 不要求額外確認並可繼續呼叫 Mailer

#### Scenario: Replay 白名單外紀錄但未重新確認
- **WHEN** record 的 From、任一收件人或 port 位於目前白名單外，且本次 replay 未提供 `--confirm-outside-whitelist`
- **THEN** 系統列出目前全部越界原因並拒絕寄送，Mailer 與 SMTP 呼叫次數皆為零

#### Scenario: Replay 白名單外紀錄並重新確認
- **WHEN** record 所有基本驗證通過且本次 replay 明確提供 `--confirm-outside-whitelist`
- **THEN** 系統只授權本次 replay 的白名單外邊界並繼續寄送

#### Scenario: Replay 確認不得略過基本驗證
- **WHEN** 本次 replay 提供確認旗標但 record 的 server、Email、port 或附件不再有效
- **THEN** 系統仍在 SMTP 前拒絕寄送，且不建立新的 history record
