## 1. 檢查與規劃

- [x] 1.1 查看 bubbles、bubbletea、lipgloss 的最新版本和 CHANGELOG
- [x] 1.2 識別可能的 breaking changes（檢查遷移指南）
- [x] 1.3 記錄當前版本號（用於對比）
- [x] 1.4 建立獨立分支進行升級

## 2. 升級依賴

- [x] 2.1 執行 `go get -u github.com/charmbracelet/lipgloss@latest`
- [x] 2.2 執行 `go get -u github.com/charmbracelet/bubbles@latest`
- [x] 2.3 執行 `go get -u github.com/charmbracelet/bubbletea@latest`
- [x] 2.4 執行 `go mod tidy` 解決間接依賴
- [x] 2.5 驗證 go.mod 和 go.sum 更新正確

## 3. 編譯與適配

- [x] 3.1 編譯代碼 `go build ./...`
- [x] 3.2 記錄所有編譯錯誤
- [x] 3.3 根據錯誤訊息更新 tui/compose.go 中的 API 調用
- [x] 3.4 根據錯誤訊息更新 tui/mail_burst.go 中的 API 調用
- [x] 3.5 根據錯誤訊息更新其他 tui/ 文件中的相關代碼
- [x] 3.6 確保代碼編譯成功，無錯誤或警告

## 4. 功能測試

- [x] 4.1 運行全部單元測試 `go test ./...`
- [ ] 4.2 手動啟動 hermes，進入 compose 頁面
- [ ] 4.3 驗證 compose 頁面：Header、Composer、Preview 邊框和輸入正常
- [ ] 4.4 測試 compose 的快捷鍵（Tab、Ctrl+J、Ctrl+K 等）
- [ ] 4.5 手動進入 mail_burst 頁面，驗證邊框和列表顯示正常
- [ ] 4.6 測試所有已知的 TUI 交互場景

## 5. 視覺檢查與驗證

- [ ] 5.1 檢查邊框樣式是否正確渲染（RoundedBorder 等）
- [ ] 5.2 檢查顏色和前景色是否如預期顯示
- [ ] 5.3 檢查文本對齐和 padding 是否正確
- [ ] 5.4 對比升級前後的 UI 截圖（如有變化，確認是否為預期）

## 6. 最終驗證與提交

- [x] 6.1 運行完整的測試套件 `go test ./...`
- [x] 6.2 執行 `go mod tidy`（確保最終狀態乾淨）
- [x] 6.3 確認沒有新增的 build warnings
- [ ] 6.4 更新 CHANGELOG 或相關文檔（如有）
- [ ] 6.5 提交 PR 並進行代碼審查
- [ ] 6.6 合併至 main 分支
