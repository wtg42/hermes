## ADDED Requirements

### Requirement: CLI 子命令直接路由至各 TUI 模式
系統 SHALL 允許使用者透過 CLI 子命令直接進入對應的 TUI 操作模式，不須經由互動式選單。

#### Scenario: 不帶子命令執行進入自訂郵件模式
- **WHEN** 使用者執行 `hermes`（不帶任何子命令）
- **THEN** 系統直接啟動「自訂郵件發送」TUI 畫面（MailFieldsModel）

#### Scenario: eml 子命令進入 eml 發送模式
- **WHEN** 使用者執行 `hermes eml`
- **THEN** 系統直接啟動 eml 檔案選擇 TUI 畫面（EmlModel）

#### Scenario: burst 子命令發送郵件
- **WHEN** 使用者執行 `hermes burst --quantity N --host H --port P`
- **THEN** 系統執行 Burst Mode 發送，行為與現有實作相同

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
