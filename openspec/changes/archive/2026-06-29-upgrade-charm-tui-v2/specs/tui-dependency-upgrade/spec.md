## MODIFIED Requirements

### Requirement: TUI 依賴升級至最新版本

系統 SHALL 將 Charm TUI framework 相關依賴升級至 latest stable v2 module path。升級後所有現有 TUI 功能 SHALL 繼續運作，無功能回歸或不可接受的視覺差異。

#### Scenario: bubbles v2 依賴升級成功

- **WHEN** 執行 `go get charm.land/bubbles/v2@latest`
- **THEN** `go.mod` SHALL 使用 `charm.land/bubbles/v2`
- **AND** 專案 SHALL 編譯成功，無編譯錯誤

#### Scenario: bubbletea v2 依賴升級成功

- **WHEN** 執行 `go get charm.land/bubbletea/v2@latest`
- **THEN** `go.mod` SHALL 使用 `charm.land/bubbletea/v2`
- **AND** 專案 SHALL 編譯成功，無編譯錯誤

#### Scenario: lipgloss v2 依賴升級成功

- **WHEN** 執行 `go get charm.land/lipgloss/v2@latest`
- **THEN** `go.mod` SHALL 使用 `charm.land/lipgloss/v2`
- **AND** 專案 SHALL 編譯成功，無編譯錯誤

#### Scenario: 舊 Charm v1 module path 已移除

- **WHEN** 升級完成並檢查 Go 原始碼與 `go.mod`
- **THEN** 系統 SHALL NOT 使用 `github.com/charmbracelet/bubbletea`
- **AND** 系統 SHALL NOT 使用 `github.com/charmbracelet/bubbles`
- **AND** 系統 SHALL NOT 使用 `github.com/charmbracelet/lipgloss`

### Requirement: 現有 TUI 功能完整運作

升級依賴後，所有現有 TUI 頁面 SHALL 完整保持功能和視覺呈現。若 Bubble Tea v2 或 Lip Gloss v2 renderer 造成微小視覺差異，該差異 MUST 經手動驗證確認不影響操作。

#### Scenario: Compose 頁面功能正常

- **WHEN** 啟動 hermes 並進入 compose 頁面
- **THEN** Header、Composer、Preview 三個面板 SHALL 顯示正常，邊框渲染正確，所有輸入和交互響應正常

#### Scenario: EML 頁面功能正常

- **WHEN** 啟動 hermes 並進入 eml 頁面
- **THEN** EML filepicker SHALL 顯示正常，檔案選取與退出快捷鍵 SHALL 正常運作

#### Scenario: 鍵盤快捷鍵運作正常

- **WHEN** 在任何 TUI 頁面按下已定義的快捷鍵（如 Ctrl+S、Ctrl+J、Ctrl+K、Ctrl+A、Tab、Shift+Tab、Esc、Ctrl+C）
- **THEN** 系統 SHALL 正確響應，焦點切換、附件選取、字段填入、發信、退出或返回等功能 SHALL 正常運作

#### Scenario: Bubble Tea v2 declarative view 適配成功

- **WHEN** TUI model 以 Bubble Tea v2 執行
- **THEN** 每個 Bubble Tea model 的 `View` SHALL 回傳 `tea.View`
- **AND** 需要 Alt Screen 的畫面 SHALL 透過 `tea.View` 宣告 Alt Screen 狀態

### Requirement: 相依依賴自動解決

升級主要依賴後，相關的間接依賴 SHALL 自動解決，無衝突或相容性問題。測試驗證 MUST 使用可寫 Go build cache，以避免 sandbox 或非互動環境造成誤判。

#### Scenario: 間接依賴自動更新

- **WHEN** 執行 `go mod tidy`
- **THEN** `go.mod` 與 `go.sum` 中的間接依賴 SHALL 自動調整至相容版本，專案 SHALL 編譯成功

#### Scenario: 全部測試通過

- **WHEN** 執行 `GOCACHE=/private/tmp/hermes-gocache go test ./...`
- **THEN** 所有單元測試 SHALL 通過，無新增失敗

#### Scenario: 升級前 baseline 測試問題已處理

- **WHEN** 執行升級前測試 baseline
- **THEN** 既有 `TestDrawLogo` 非互動 terminal 失敗 SHALL 已修正、隔離或明確記錄
- **AND** 升級後測試失敗 SHALL 可歸因於 v2 遷移本身，而不是既有 baseline 問題
