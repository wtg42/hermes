## 1. Baseline 與環境確認

- [x] 1.1 確認 Codex shell 可直接執行 `go version`，且版本為 Go 1.26.x 或更新。
- [x] 1.2 使用可寫 cache 執行 `GOCACHE=/private/tmp/hermes-gocache go test ./...`，記錄升級前 baseline。
- [x] 1.3 修正或隔離 `TestDrawLogo` 在非互動 terminal 下的既有失敗，使升級前測試 baseline 可判讀。
- [x] 1.4 重新執行 `GOCACHE=/private/tmp/hermes-gocache go test ./...`，確認 baseline 不再因既有問題失敗。

## 2. Charm TUI v2 依賴遷移

- [x] 2.1 將 Go imports 從 `github.com/charmbracelet/bubbletea` 改為 `charm.land/bubbletea/v2`。
- [x] 2.2 將 Go imports 從 `github.com/charmbracelet/bubbles/...` 改為 `charm.land/bubbles/v2/...`。
- [x] 2.3 將 Go imports 從 `github.com/charmbracelet/lipgloss` 改為 `charm.land/lipgloss/v2`。
- [x] 2.4 執行 `go get charm.land/bubbletea/v2@latest charm.land/bubbles/v2@latest charm.land/lipgloss/v2@latest`。
- [x] 2.5 執行 `go mod tidy`，確認 `go.mod` 與 `go.sum` 僅包含預期的 Charm TUI stack v2 變更與相容間接依賴更新。

## 3. Bubble Tea v2 API 適配

- [x] 3.1 將 `tui/compose.go` 的 `ComposeModel.View() string` 改為 `View() tea.View`，保留既有 render content。
- [x] 3.2 將 `tui/load_eml.go` 的 `EmlModel.View() string` 改為 `View() tea.View`，保留既有 render content。
- [x] 3.3 將 `tui/alert.go` 的 `AlertModel.View() string` 改為 `View() tea.View`，保留既有 render content。
- [x] 3.4 移除 `cmd/root_cmd.go` 與 `cmd/eml_cmd.go` 的 `tea.WithAltScreen()`，改由相關 model 的 `tea.View` 設定 `AltScreen = true`。
- [x] 3.5 將 `tea.KeyMsg` 一般 key press 處理改為 `tea.KeyPressMsg`，並更新 `handleHeaderKeys`、`handleComposerKeys` 等 helper 參數型別。
- [x] 3.6 檢查 `tea.ClearScreen`、`tea.WindowSizeMsg`、`tea.Cmd`、`tea.Model` 等用法在 Bubble Tea v2 下是否仍相容，必要時依 v2 API 調整。
- [x] 3.7 更新 `tui/compose_test.go` 與其他受影響測試，使其使用 v2 imports 與 v2 view contract。

## 4. 自動驗證

- [x] 4.1 執行 `gofmt` 或 `go fmt ./...`，確認格式化完成。
- [x] 4.2 執行 `GOCACHE=/private/tmp/hermes-gocache go test ./...`，確認所有單元測試通過。
- [x] 4.3 執行 `make build`，確認 hermes 可建置。
- [x] 4.4 執行 `make lint`，確認 `go vet ./...` 與格式化檢查無錯誤。
- [x] 4.5 使用 `rg` 確認原始碼與 `go.mod` 不再引用 `github.com/charmbracelet/bubbletea`、`github.com/charmbracelet/bubbles`、`github.com/charmbracelet/lipgloss`。

## 5. 手動 Smoke Test

- [x] 5.1 啟動 hermes compose TUI，確認 Header、Composer、Preview 三個面板顯示與邊框正常。
- [x] 5.2 在 compose TUI 驗證 `Tab`、`Shift+Tab`、`Ctrl+J`、`Ctrl+K` 的焦點切換。
- [x] 5.3 在 compose TUI 驗證 `Ctrl+A` filepicker overlay 可開啟、選取或關閉。
- [x] 5.4 在 compose TUI 驗證 `Ctrl+S` 發信流程與 sending 狀態顯示。
- [x] 5.5 在 alert 畫面驗證 `Esc` 可返回保存的 compose model，`Ctrl+C` 可退出。
- [x] 5.6 啟動 eml TUI，確認 filepicker 顯示、檔案選取與 `esc`/`q`/`ctrl+c` 退出正常。

## 6. 收尾

- [x] 6.1 檢查 `git diff`，確認沒有無關格式化或 dependency churn。
- [x] 6.2 更新測試證據摘要，包含 Go 版本、測試指令、手動 smoke test 結果。
- [x] 6.3 若 v2 renderer 造成任何視覺差異，記錄差異與是否可接受。

## Verification Evidence

- Go 版本：`go version go1.26.4 darwin/arm64`。
- Baseline 修正後測試：`GOCACHE=/private/tmp/hermes-gocache go test ./...` 通過。
- 最終測試：`GOCACHE=/private/tmp/hermes-gocache go test ./...` 通過。
- 建置：`GOCACHE=/private/tmp/hermes-gocache make build` 通過。
- Lint/格式化：`GOCACHE=/private/tmp/hermes-gocache make lint` 通過。
- 舊 import 掃描：`rg` 確認 Go 原始碼與 `go.mod` 不再引用 `github.com/charmbracelet/bubbletea`、`github.com/charmbracelet/bubbles`、`github.com/charmbracelet/lipgloss`。
- Diff 檢查：變更集中於 Charm TUI stack v2 遷移、Bubble Tea v2 API 適配、`TestDrawLogo` baseline 修正、OpenSpec artifacts；`mail/compose.go` 只有 `go fmt ./...` 產生的欄位註解對齊。
- Compose smoke test：`TERM=xterm-256color ./bin/hermes` 可顯示 Header、Composer、Preview；`Tab`、`Shift+Tab`、`Ctrl+J`、`Ctrl+K`、`Ctrl+A` overlay、`Ctrl+S` alert、`Esc` 返回、`Ctrl+C` 退出皆已驗證。
- EML smoke test：`TERM=xterm-256color ./bin/hermes eml` 可顯示 filepicker 狀態；`q` 可退出。
- v2 renderer 差異：Lip Gloss v2 framed block width/height 語意與 v1 不同，已在 `renderHeaderPanel`、`renderComposerPanel` 與相關測試中調整；人工 smoke test 確認 compose layout 可接受。
