## ADDED Requirements

### Requirement: History structured CLI routing
系統 SHALL 將 `hermes history list`、`hermes history show <id>` 與 `hermes history replay <id>` 路由至本機 structured history 流程，不啟動 Compose TUI、EML picker 或 Burst mode。

#### Scenario: 執行 history list
- **WHEN** 使用者執行 `hermes history list`
- **THEN** 系統只讀取並輸出本機 history summaries，不發送郵件

#### Scenario: 執行 history show
- **WHEN** 使用者執行 `hermes history show <id>`
- **THEN** 系統只讀取並輸出指定 record，不發送郵件

#### Scenario: 執行 history replay
- **WHEN** 使用者執行帶有效 ID 與必要安全 flags 的 `hermes history replay <id>`
- **THEN** 系統只執行一次 structured 單封 replay 並在完成後結束

#### Scenario: 顯示 history 說明
- **WHEN** 使用者執行 `hermes history --help` 或任一 history 子命令的 `--help`
- **THEN** 系統列出對應參數，不啟動 TUI、讀寫 history 或發送郵件
