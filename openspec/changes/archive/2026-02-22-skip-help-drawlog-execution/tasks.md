## 1. 修改 Execute() 函式

- [x] 1.1 修改 `cmd/root_cmd.go` 中 `Execute()` 的簽名，返回型別改為 `bool`
- [x] 1.2 在 `Execute()` 中檢查是否執行了幫助命令（檢查 `cmd.Name()` 或 flag 狀態）
- [x] 1.3 當幫助被顯示時返回 `false`，否則返回 `true`

## 2. 修改 main() 邏輯

- [x] 2.1 修改 `main.go` 中 `cmd.Execute()` 的呼叫，接收返回值
- [x] 2.2 根據返回值決定是否執行 `drawLogo()`
- [x] 2.3 確保錯誤情況下的行為正確（如 Execute 返回錯誤）

## 3. 測試驗證

- [x] 3.1 執行 `hermes -h` 確認幫助顯示且無 gopher 圖像
- [x] 3.2 執行 `hermes --help` 確認幫助顯示且無 gopher 圖像
- [x] 3.3 執行 `hermes burst -h` 確認子命令幫助正常且無 gopher 圖像
- [x] 3.4 執行 `hermes` 或其他有效命令確認 gopher 圖像正常顯示
- [x] 3.5 執行 `go test ./cmd -v` 確保現有單元測試通過

## 4. 程式碼品質檢查

- [x] 4.1 執行 `go fmt ./...` 確保代碼格式正確
- [x] 4.2 執行 `go vet ./...` 檢查潛在問題
- [x] 4.3 運行完整測試 `make test` 確保無回歸
