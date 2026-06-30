## Why

Hermes 目前仍使用 Charmbracelet TUI stack 的 v1 module path 與 Bubble Tea v1 API；Bubble Tea v2 已改為 `charm.land/.../v2` module path，並引入 declarative `tea.View`、新的 key/mouse event model 與 renderer 行為。若要持續取得最新穩定版的修正與效能改善，需要以一次完整遷移處理 dependency、API 與回歸驗證，而不是只更新版本號。

## What Changes

- **BREAKING**：將 TUI 相關依賴從 `github.com/charmbracelet/...` v1 module path 遷移至 `charm.land/.../v2`：
  - `charm.land/bubbletea/v2`
  - `charm.land/bubbles/v2`
  - `charm.land/lipgloss/v2`
- **BREAKING**：將 Bubble Tea model 的 `View() string` 遷移為 `View() tea.View`，並在 view 中宣告 Alt Screen 等 terminal state。
- **BREAKING**：將一般鍵盤事件處理從 `tea.KeyMsg` struct 用法遷移至 Bubble Tea v2 的 `tea.KeyPressMsg`。
- 將 `tea.WithAltScreen()` 等 v1 imperative program options 移除，改由 `tea.View` fields 宣告。
- 將既有 TUI 元件 import 與測試更新為 Bubbles/Lip Gloss v2 相容用法。
- 在升級前處理或隔離既有 `TestDrawLogo` 非互動 terminal baseline 問題，避免升級後測試失敗歸因不清。
- 驗證升級後 compose 與 eml TUI 的視覺、快捷鍵、filepicker、alert 返回流程維持既有行為。

## Capabilities

### New Capabilities

- 無。

### Modified Capabilities

- `tui-dependency-upgrade`：將「升級至最新穩定版」的需求明確更新為 Charm TUI stack v2 module path 遷移、Bubble Tea v2 API 適配、Go toolchain 相容性與回歸驗證。

## Impact

- 影響 dependency：`go.mod`、`go.sum` 中的 Bubble Tea、Bubbles、Lip Gloss 與相關間接依賴。
- 影響 TUI API 使用：
  - `cmd/root_cmd.go`
  - `cmd/eml_cmd.go`
  - `main.go`
  - `tui/alert.go`
  - `tui/load_eml.go`
  - `tui/compose.go`
  - `tui/components.go`
  - `tui/compose_test.go`
- 影響驗證流程：需要使用 Go 1.26 可用環境，並在 sandbox 中以可寫 `GOCACHE` 執行測試。
- 使用者可見行為原則上不變；任何視覺差異都必須來自 Bubble Tea/Lip Gloss v2 renderer 或樣式行為差異，且需經手動 smoke test 確認可接受。
