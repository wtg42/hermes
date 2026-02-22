## Why

目前 Hermes 的 TUI 依賴版本較舊（bubbles v1.0.0、bubbletea v1.3.10、lipgloss v1.1.0）。升級到最新版本可以獲得性能改進、新功能、錯誤修復，以及與上游生態的最佳實踐保持一致。特別是 lipgloss 最新版本提供了 Compositing API 和更強大的邊框自訂能力，為未來的 UI 增強奠定基礎。

## What Changes

- 升級 `github.com/charmbracelet/bubbles` 到最新穩定版本
- 升級 `github.com/charmbracelet/bubbletea` 到最新穩定版本
- 升級 `github.com/charmbracelet/lipgloss` 到最新穩定版本（或 v2.0-beta，如果穩定性允許）
- 解決任何 breaking changes，確保現有 TUI 功能完整運作
- 更新相關的間接依賴以確保兼容性

## Capabilities

### New Capabilities

- `tui-dependency-upgrade`: 統一升級 TUI 框架依賴，支援最新的邊框自訂、渲染優化和佈局能力

### Modified Capabilities

- 無現有功能需要變更需求（implementation details 可能更新）

## Impact

- **影響的代碼**：`tui/` 目錄下的所有 TUI 組件和邊框渲染邏輯
- **相依性**：go.mod 中的三個主要 Charmbracelet 套件
- **測試**：需要運行全部 TUI 相關測試，確保 compose、mail_burst 等頁面運作正常
- **可能的 Breaking Changes**：需要檢查新版本的 API 變更，適配相關代碼
