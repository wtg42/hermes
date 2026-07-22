## MODIFIED Requirements

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
