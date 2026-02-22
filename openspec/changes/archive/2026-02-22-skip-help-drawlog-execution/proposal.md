## Why

當使用者執行 `hermes -h` 或 `hermes --help` 查看幫助文件時，程式會繪製並輸出 ASCII art gopher 圖像。這是不必要的資源浪費，降低了幫助功能的使用體驗（增加延遲）。應該只在程式正常執行完成時才繪製圖像。

## What Changes

- 修改 `cmd.Execute()` 返回一個布爾值，指示是否顯示過幫助訊息
- 修改 `main()` 根據 `Execute()` 的返回值決定是否執行 `drawLogo()`
- 幫助訊息顯示時，直接跳過 gopher 圖像繪製

## Capabilities

### New Capabilities

- `help-skip-drawlogo`: 實現幫助訊息邏輯分離，不在幫助時執行非必要的圖像繪製操作

### Modified Capabilities

<!-- 無現有功能需要修改需求 -->

## Impact

- **影響的代碼**：`main.go`、`cmd/root_cmd.go`
- **API 變更**：`Execute()` 函式新增返回值（布爾值）
- **使用者體驗**：幫助指令響應速度提升，避免不必要的輸出延遲
