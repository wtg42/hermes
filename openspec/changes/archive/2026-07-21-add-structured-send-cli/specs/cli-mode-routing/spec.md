## ADDED Requirements

### Requirement: Structured send CLI routing
系統 SHALL 將 `hermes send` 直接路由到 structured 單封寄信流程，不啟動 TUI、EML picker 或 Burst mode。

#### Scenario: 執行 send 子命令
- **WHEN** 使用者執行帶有效 structured flags 的 `hermes send`
- **THEN** 系統只執行一次 structured 單封寄信並在完成後結束

#### Scenario: 顯示 send 說明
- **WHEN** 使用者執行 `hermes send --help`
- **THEN** 系統列出 structured send flags，不啟動 TUI 或發送郵件
