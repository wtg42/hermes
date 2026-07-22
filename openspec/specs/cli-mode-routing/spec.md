# cli-mode-routing Specification

## Purpose
TBD - created by archiving change remove-menu-use-cli-args. Update Purpose after archive.
## Requirements
### Requirement: CLI 子命令直接路由至各 TUI 模式
系統 SHALL 允許使用者透過 CLI 子命令直接進入對應的 TUI 操作模式，不須經由互動式選單；Burst mode SHALL 先完成地址與網域安全驗證，驗證失敗時命令 MUST 回傳錯誤且不得寄信。

#### Scenario: 不帶子命令執行進入自訂郵件模式
- **WHEN** 使用者執行 `hermes`（不帶任何子命令）
- **THEN** 系統直接啟動「自訂郵件發送」TUI 畫面（MailFieldsModel）

#### Scenario: eml 子命令進入 eml 發送模式
- **WHEN** 使用者執行 `hermes eml`
- **THEN** 系統直接啟動 eml 檔案選擇 TUI 畫面（EmlModel）

#### Scenario: burst 子命令使用安全隨機網域發送郵件
- **WHEN** 使用者執行 `hermes burst --quantity N --host H --port P --domain rd01.softnext.com.tw`
- **THEN** 系統通過安全驗證並執行 Burst mode 併發發送

#### Scenario: burst 子命令使用固定地址發送郵件
- **WHEN** 使用者執行 `hermes burst --quantity N --host H --port P --from F --to T`，且 F 與 T 的網域都在正向表列或已明確授權
- **THEN** 系統不要求 `--domain`，並以固定 From 與 To 執行 Burst mode 發送

#### Scenario: burst 子命令缺少隨機網域
- **WHEN** 使用者執行 `hermes burst --quantity N --host H --port P`，未指定固定 From、固定 To 或 `--domain`
- **THEN** 系統回傳缺少隨機地址網域的錯誤，且不送出任何郵件

### Requirement: Logo 於程式結束後顯示
系統 SHALL 在 TUI 程式結束後才繪製 gopher logo，避免 logo 干擾 TUI 畫面。

#### Scenario: 正常退出後顯示 logo
- **WHEN** 使用者在任一 TUI 模式中退出程式
- **THEN** 系統於終端機輸出 gopher logo 後結束

### Requirement: 移除 TUI 選單入口
系統 SHALL NOT 提供 `start-tui` 子命令或任何互動式選單作為模式選擇入口。

#### Scenario: 執行已移除的 start-tui 子命令
- **WHEN** 使用者執行 `hermes start-tui`
- **THEN** 系統回傳 unknown command 錯誤並顯示說明

### Requirement: Structured send CLI routing
系統 SHALL 將 `hermes send` 直接路由到 structured 單封寄信流程，不啟動 TUI、EML picker 或 Burst mode。

#### Scenario: 執行 send 子命令
- **WHEN** 使用者執行帶有效 structured flags 的 `hermes send`
- **THEN** 系統只執行一次 structured 單封寄信並在完成後結束

#### Scenario: 顯示 send 說明
- **WHEN** 使用者執行 `hermes send --help`
- **THEN** 系統列出 structured send flags，不啟動 TUI 或發送郵件

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
