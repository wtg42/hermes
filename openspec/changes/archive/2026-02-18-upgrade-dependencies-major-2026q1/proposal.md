## Why

目前專案依賴版本已有一段時間未更新，存在安全性修補落後、相容性風險累積與後續升級成本變高的問題。此變更將以一次性 major 升級整理依賴基線，降低技術債並確保開發與 CI 能持續跟進。

## What Changes

- 升級核心直接依賴到目前穩定版本，包含 `github.com/charmbracelet/bubbles`、`github.com/charmbracelet/bubbletea`、`github.com/spf13/cobra`、`github.com/spf13/viper`、`github.com/stretchr/testify`、`github.com/samber/lo`、`golang.org/x/term`。
- 更新 `go.mod` 與 `go.sum`，同步整理由直接依賴引發的 indirect 版本變更。
- 針對 TUI 互動流程進行回歸驗證，涵蓋輸入欄位、捲動元件、檔案挑選與關鍵快捷操作。
- 明確定義升級完成標準：`make lint`、`make test`、`make build` 皆須通過，且主要 CLI/TUI 流程可正常使用。
- 若遇到 major 變更造成介面或行為不相容，補齊必要調整與測試案例以維持既有使用體驗。

## Capabilities

### New Capabilities
- `dependency-major-upgrade`: 定義 Hermes 進行依賴 major 升級時的行為、驗證門檻與回歸範圍。

### Modified Capabilities
- 無

## Impact

- Affected code: `go.mod`、`go.sum`、`cmd/`、`tui/`、`utils/`、`sendmail/`（依實際相容性調整而定）。
- Affected tooling: Go toolchain 與 CI 執行環境需可支援升級後依賴需求。
- Risk areas: TUI 元件 API 相容性（特別是 `bubbles` major 版本），以及事件處理行為差異。
