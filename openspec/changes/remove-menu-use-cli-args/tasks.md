## 1. 移除舊有 menu 與 tui_cmd

- [x] 1.1 刪除 `tui/menu.go`
- [x] 1.2 刪除 `tui/menu_test.go`
- [x] 1.3 刪除 `cmd/tui_cmd.go`
- [x] 1.4 刪除 `cmd/tui_cmd_test.go`

## 2. 修改 root_cmd 作為預設進入點

- [x] 2.1 移除 `rootCmd.AddCommand(startTUICmd)`
- [x] 2.2 在 `rootCmd` 加入 `Run`，直接啟動 `MailFieldsModel` bubbletea program

## 3. 新增 eml 子命令

- [x] 3.1 新增 `cmd/eml_cmd.go`，實作 `hermes eml` 子命令
- [x] 3.2 在 `eml_cmd.go` 的 `Run` 中初始化 `EmlModel`，呼叫 `filepicker.Init()` 取得初始 cmd，啟動 bubbletea program
- [x] 3.3 在 `root_cmd.go` 的 `init()` 中加入 `rootCmd.AddCommand(emlCmd)`

## 4. 調整 main.go logo 時機

- [x] 4.1 將 `drawLogo()` 呼叫從 `cmd.Execute()` 之前移至之後
