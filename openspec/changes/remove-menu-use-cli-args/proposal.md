## Why

目前使用者必須先執行 `hermes start-tui` 進入互動式選單，才能選擇要使用的模式，多了一層不必要的選擇步驟。將模式選擇移至 CLI 參數層，讓使用者可以直接進入目標模式，減少操作步驟。

## What Changes

- **BREAKING** 移除 `start-tui` 子命令
- `hermes`（不帶子命令）預設直接啟動「自訂郵件發送」TUI 畫面
- 新增 `hermes eml` 子命令，直接啟動 eml 檔案選擇 TUI 畫面
- `hermes burst` 子命令維持不變
- 移除 `tui/menu.go`（menu model 及 `StartMenu` function）
- 移除 `cmd/tui_cmd.go` 及對應測試
- 移除 `tui/menu_test.go`
- 調整 `main.go`：logo 改為在程式結束後才繪製（執行完 TUI 後印出）

## Capabilities

### New Capabilities

- `cli-mode-routing`：透過 CLI 子命令直接路由至各 TUI 模式，不經由 menu 選擇畫面

### Modified Capabilities

（無 spec 層級的需求變更，`unified-email-sending` 的行為不受影響）

## Impact

- `main.go`：調整 logo 繪製時機
- `cmd/root_cmd.go`：加入 `Run` 直接啟動 MailFieldsModel
- `cmd/tui_cmd.go`：刪除
- `cmd/tui_cmd_test.go`：刪除
- `cmd/eml_cmd.go`：新增
- `tui/menu.go`：刪除
- `tui/menu_test.go`：刪除
