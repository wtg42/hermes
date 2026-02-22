## Purpose

Ensure that the TUI framework dependencies (lipgloss, bubbles, bubbletea) are kept up-to-date with the latest stable versions, providing users with the latest performance improvements, features, and security patches.

## Requirements

### Requirement: TUI 依賴升級至最新版本

系統應升級 Charmbracelet TUI 框架相關依賴至最新穩定版本。升級後所有現有 TUI 功能應繼續運作，無功能回歸或視覺差異（除非版本更新有意改變）。

#### Scenario: bubbles 依賴升級成功

- **WHEN** 執行 `go get -u github.com/charmbracelet/bubbles@latest`
- **THEN** go.mod 中 bubbles 版本更新，編譯成功，無編譯錯誤

#### Scenario: bubbletea 依賴升級成功

- **WHEN** 執行 `go get -u github.com/charmbracelet/bubbletea@latest`
- **THEN** go.mod 中 bubbletea 版本更新，編譯成功，無編譯錯誤

#### Scenario: lipgloss 依賴升級成功

- **WHEN** 執行 `go get -u github.com/charmbracelet/lipgloss@latest`
- **THEN** go.mod 中 lipgloss 版本更新，編譯成功，無編譯錯誤

### Requirement: 現有 TUI 功能完整運作

升級依賴後，所有現有 TUI 頁面應完整保持功能和視覺呈現。

#### Scenario: Compose 頁面功能正常

- **WHEN** 啟動 hermes 並進入 compose 頁面
- **THEN** Header、Composer、Preview 三個面板顯示正常，邊框渲染正確，所有輸入和交互響應正常

#### Scenario: Mail Burst 頁面功能正常

- **WHEN** 啟動 hermes 並進入 mail_burst 頁面
- **THEN** 郵件列表、詳情面板顯示正常，所有邊框和樣式渲染正確

#### Scenario: 鍵盤快捷鍵運作正常

- **WHEN** 在任何 TUI 頁面按下已定義的快捷鍵（如 Ctrl+S、Ctrl+J、Tab 等）
- **THEN** 系統正確響應，焦點切換、字段填入、發信等功能運作

### Requirement: 相依依賴自動解決

升級主要依賴後，相關的間接依賴應自動解決，無衝突或兼容性問題。

#### Scenario: 間接依賴自動更新

- **WHEN** 執行 `go mod tidy`
- **THEN** go.mod 中的間接依賴自動調整至兼容版本，編譯成功

#### Scenario: 全部測試通過

- **WHEN** 執行 `go test ./...`
- **THEN** 所有單元測試通過，無新增失敗
