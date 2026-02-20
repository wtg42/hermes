## Context

hermes 目前的進入點是 `hermes start-tui`，執行後會啟動一個 bubbletea menu 讓使用者選擇模式（自訂郵件、Burst、eml）。這個 menu 是多餘的間接層，使用者每次都要額外按一個按鍵才能進入目標模式。

此設計將模式選擇移至 CLI 子命令層，移除 menu 這層。

## Goals / Non-Goals

**Goals:**
- `hermes` 不帶子命令直接啟動「自訂郵件發送」TUI 畫面
- 新增 `hermes eml` 子命令直接啟動 eml TUI 畫面
- `hermes burst` 維持不變
- logo 移至 TUI 結束後才繪製

**Non-Goals:**
- 不改變各 TUI 畫面本身的互動邏輯
- 不修改 `sendmail` 層的任何邏輯
- 不新增任何 flag 或 config

## Decisions

### 決策 1：rootCmd 加上 Run 作為預設進入點

`rootCmd` 直接加上 `Run` function，執行 `tui.InitialMailFieldsModel()` 啟動 bubbletea program。

**替代方案考量**：新增 `hermes send` 子命令。
**選擇原因**：自訂郵件是主要用途，不帶參數直接進是最低摩擦的體驗。

### 決策 2：新增 eml 子命令（`cmd/eml_cmd.go`）

直接呼叫 `tui.InitialEmlModel()` 並執行其 `Init()` 取得初始 cmd，然後啟動 bubbletea program。

**注意**：EmlModel 需要先呼叫 `filepicker.Init()` 取得初始 cmd，這個行為必須在 cmd 層正確處理。

### 決策 3：移除 `tui/menu.go` 及 `StartMenu`

menu 的唯一職責是路由，這個職責移至 CLI 層後，menu 沒有存在意義。`StartMenu` var function 僅用於測試替換，一併移除。

### 決策 4：logo 改為結束後繪製

`main.go` 將 `drawLogo()` 呼叫移至 `cmd.Execute()` 之後，讓 TUI 畫面不被 logo 干擾，使用者退出後才看到 logo。

## Risks / Trade-offs

- **BREAKING 變更** `start-tui` 子命令消失 → 有使用 `hermes start-tui` 的腳本或習慣需要更新
- EmlModel 初始化需要終端環境（filepicker 依賴），CI 測試需 Skip，與現有做法一致

## Migration Plan

1. 刪除 `cmd/tui_cmd.go`、`cmd/tui_cmd_test.go`
2. 刪除 `tui/menu.go`、`tui/menu_test.go`
3. 修改 `cmd/root_cmd.go`：加 Run、移除 `startTUICmd`
4. 新增 `cmd/eml_cmd.go`
5. 修改 `main.go`：logo 移到 Execute() 之後

回滾：git revert 即可，無資料或外部狀態變更。
