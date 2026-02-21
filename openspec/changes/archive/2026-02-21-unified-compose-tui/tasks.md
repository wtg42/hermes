## 1. 建立 ComposeModel 基礎架構

- [x] 1.1 建立 `tui/compose.go` 檔案，定義 `ComposeModel` struct（包括 mailFields、composer、preview、activePanel 等欄位）
- [x] 1.2 實作 `InitialComposeModel()` 函數，初始化 7 個 textinput、textarea、viewport
- [x] 1.3 複用 `MailFieldsModel` 的 textinput 初始化邏輯，設定 placeholder 與 focus 樣式

## 2. Update 函數與按鍵路由

- [x] 2.1 實作全域快捷鍵：`ctrl+c` (Quit)、`esc` (Clear/Quit)
- [x] 2.2 實作 Header panel 的按鍵路由：`tab` / `shift+tab` 在 textinput 間導航、`ctrl+j` 切換到 Composer
- [x] 2.3 實作 Composer panel 的按鍵路由：`ctrl+k` 回到 Header、`ctrl+h/t/e` 填入範本、`ctrl+a` 觸發附件選取
- [x] 2.4 實作 `ctrl+s` 觸發發信（複用 `sendMailWithChannel` 邏輯）

## 3. Preview 功能實現

- [x] 3.1 初始化 `viewport.Model` 作為 Preview panel
- [x] 3.2 在 Composer textarea 更新後，透過 `preview.SetContent(m.composer.Value())` 同步內容
- [x] 3.3 確保 Preview viewport 可捲動（超過視窗高度時顯示捲軸）

## 4. Filepicker Overlay 附件選取

- [x] 4.1 新增 `showFilePicker bool` 和 `filepicker filepicker.Model` 欄位到 ComposeModel
- [x] 4.2 實作 Overlay 觸發邏輯：`ctrl+a` 或狀態列 `[📎Attach]` 按鈕打開 filepicker
- [x] 4.3 處理 filepicker 的選擇結果：選擇後返回撰寫畫面，焦點復位
- [x] 4.4 實作 Overlay 時的按鍵路由隔離（filepicker 獨立路由，不影響撰寫畫面按鍵）

## 5. View 函數與視覺設計

- [x] 5.1 使用 `lipgloss.JoinHorizontal` 實現左右 50:50 分割佈局
- [x] 5.2 左側：垂直分割 Header panel（上 40%）+ Composer panel（下 60%）
- [x] 5.3 右側：Preview panel（viewport 顯示）
- [x] 5.4 根據 `activePanel` 改變焦點 panel 的 border 顏色（橘色 `#DC851C` vs 灰色）
- [x] 5.5 實作底部狀態列：顯示 `[⚡Send] [📎Attach] [F3→Quit] [Ctrl+C]` + SMTP 連線資訊
- [x] 5.6 在 Header/Composer border title 中加入 `...` 和 `▽` 作為視覺裝飾

## 6. Overlay 與視窗管理

- [x] 6.1 實作 Filepicker Overlay 的渲染邏輯：當 `showFilePicker == true` 時，以 overlay 覆蓋 Composer 區域
- [x] 6.2 處理 Overlay 時的視窗大小同步（`tea.WindowSizeMsg` 更新時重新計算 panel 寬高）
- [x] 6.3 確保 Overlay 關閉後焦點正確復位

## 7. 資料流與發信整合

- [x] 7.1 在發信前，將所有 Header 欄位值寫入 viper（複用 `MailFieldsModel` 的邏輯）
- [x] 7.2 在發信前，將 Composer 內容寫入 viper（複用 `MailMsgModel` 的邏輯）
- [x] 7.3 呼叫現有的 `sendmail.SendMailWithMultipart` 函數發信
- [x] 7.4 發信完成後顯示 `AlertModel` 結果提示框（複用舊設計）

## 8. 更新入口點

- [x] 8.1 修改 `cmd/root_cmd.go`，將 `InitialMailFieldsModel()` 改為 `InitialComposeModel()`
- [x] 8.2 確保 `main.go` 與 cobra 命令初始化邏輯不受影響

## 9. 測試與驗證

- [x] 9.1 編譯檢查：`go build ./...` 無誤
- [ ] 9.2 手動測試：執行 `hermes`，確認顯示新的分割撰寫畫面
- [ ] 9.3 測試焦點切換：`Ctrl+J` / `Ctrl+K` 在 Header ↔ Composer 間切換
- [ ] 9.4 測試 Preview 同步：在 Composer 輸入文字，右側 Preview 即時更新
- [ ] 9.5 測試快捷鍵：Tab/Shift+Tab、Ctrl+H/T/E、Ctrl+A、Ctrl+S 等
- [ ] 9.6 測試發信流程：填入有效 SMTP 資訊，驗證發信成功與 AlertModel 顯示
- [ ] 9.7 測試 Filepicker Overlay：打開、選擇檔案、取消等操作
- [x] 9.8 執行現有測試：`go test ./...` 確認不破壞舊有功能
- [ ] 9.9 邊界測試：終端機寬度不足時的視覺表現、超大內文的 Preview 捲動

## 10. 清理與文檔

- [x] 10.1 確認舊的 `MailFieldsModel` 和 `MailMsgModel` 在程式碼中不再被主流程使用
- [x] 10.2 檢查是否需要更新 README 或使用說明（如有）
- [x] 10.3 驗證編譯與執行無警告
