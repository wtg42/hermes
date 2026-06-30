## Context

Hermes 的 TUI 目前使用 Bubble Tea v1.3.10、Bubbles v1.0.0、Lip Gloss v1.1.0，module path 仍是 `github.com/charmbracelet/...`。Bubble Tea v2 與相鄰套件已改為 `charm.land/.../v2`，且 Bubble Tea model contract、terminal option 管理、keyboard message 型別都有 breaking changes。

這次升級會跨越 `go.mod`、CLI 啟動點與所有 Bubble Tea model。專案已確認 Codex 環境可直接使用 Go 1.26.4，因此 latest v2 stack 的 Go 1.25+ 門檻不是 blocker。不過目前 `go test ./...` 在 sandbox 內需使用可寫 `GOCACHE`，且 root package 的 `TestDrawLogo` 在非互動 terminal 中已有 baseline 失敗，升級前需先處理或隔離。

## Goals / Non-Goals

**Goals:**

- 將 Bubble Tea、Bubbles、Lip Gloss 升級至 latest v2 stable module path。
- 適配 Bubble Tea v2 的 `tea.View` declarative API。
- 適配 Bubble Tea v2 的 key press event model。
- 保持既有 compose 與 eml TUI 使用者行為不變。
- 建立可判讀的升級前後測試 baseline，避免既有 `TestDrawLogo` 問題污染 v2 regression 判斷。

**Non-Goals:**

- 不重新設計 TUI layout、快捷鍵或郵件撰寫流程。
- 不導入 Bubble Tea v2 新功能作為產品功能，例如 native progress bar、clipboard、terminal color control。
- 不變更 CLI command surface 或 SMTP 發信邏輯。
- 不把這次升級擴大成所有 Go dependencies 的全面 major upgrade。

## Decisions

### Decision: 以 latest v2 stack 一次遷移三個 Charm 套件

選擇同時升級 `charm.land/bubbletea/v2`、`charm.land/bubbles/v2`、`charm.land/lipgloss/v2`，而不是只升 Bubble Tea。

理由：Hermes 的 TUI model 同時使用 Bubble Tea runtime、Bubbles components 與 Lip Gloss styling；Bubbles v2 本身依賴 Bubble Tea v2 與 Lip Gloss v2。分批升級會產生混用 v1/v2 module path 的風險，增加 import 與 type mismatch 的機率。

替代方案：先只升 Bubble Tea v2。此方案看似小，但 Bubbles/Lip Gloss 仍會停在舊 module path，容易造成相容層複雜化，因此不採用。

### Decision: 將 Alt Screen 狀態放入各 model 的 `View() tea.View`

Bubble Tea v2 移除 `tea.WithAltScreen()` 這類 imperative program options，改由 view fields 宣告 terminal state。Hermes 目前 root compose 與 eml mode 都使用 Alt Screen，因此各 TUI model 的 `View()` 需要回傳 `tea.View` 並設定 `AltScreen = true`。

替代方案：新增 wrapper model 統一包裝 Alt Screen。此方案會讓目前簡單的 model 結構多一層間接性，且沒有明顯收益，因此優先使用各 model 直接宣告。

### Decision: 一般鍵盤處理改用 `tea.KeyPressMsg`

Hermes 目前快捷鍵都只處理 key press，不需要 key release。Bubble Tea v2 中 `tea.KeyMsg` 是 key press/release 的 interface，因此一般分支應改為 `tea.KeyPressMsg`，並保留既有 `msg.String()` 快捷鍵比對。

替代方案：繼續 match `tea.KeyMsg` 後再 type switch。此方案適合需要同時處理 release 的 app，但 Hermes 目前不需要，會讓程式更囉嗦。

### Decision: 先解決測試 baseline，再做 dependency/API 遷移

目前 `GOCACHE=/private/tmp/hermes-gocache go test ./...` 可執行，但 root package 的 `TestDrawLogo` 在非互動 terminal 中因 color capability 偵測失敗。此問題不是 v2 升級造成，應在 dependency bump 前先修正、隔離或明確標記，使後續失敗能歸因於 v2 變更。

替代方案：直接升級後再處理所有測試失敗。此方案會混淆既有測試問題與 v2 regression，不利於回歸判斷。

## Risks / Trade-offs

- [Risk] Bubble Tea v2 renderer 改變可能造成邊框、寬度、游標或清屏行為差異。→ Mitigation：保留既有 layout 測試，並新增或執行 compose/eml 手動 smoke test。
- [Risk] Bubbles v2 component API 或 key handling 行為差異導致 textinput、textarea、viewport、filepicker 交互異常。→ Mitigation：逐一驗證 Tab/Shift+Tab、Ctrl+J/Ctrl+K、Ctrl+A filepicker、Esc 與 Ctrl+C。
- [Risk] `tea.ClearScreen` 或 alert 返回流程在 v2 下行為不同。→ Mitigation：保留 alert model 返回 compose 的情境測試與手動驗證。
- [Risk] Go module tidy 可能大幅更新間接依賴。→ Mitigation：限制變更主題為 Charm TUI stack，檢查 `go.mod`/`go.sum` diff，避免混入無關 dependency upgrade。
- [Risk] sandbox 中預設 Go cache 不可寫。→ Mitigation：驗證指令使用 workspace/tmp 可寫 `GOCACHE`，並在測試證據中記錄。

## Migration Plan

1. 修正或隔離 `TestDrawLogo` 在非互動 terminal 的 baseline 失敗。
2. 將 import path 從 `github.com/charmbracelet/...` 遷移至 `charm.land/.../v2`。
3. 執行 `go get` 升級 Bubble Tea、Bubbles、Lip Gloss 至 latest v2 stable。
4. 將 `View() string` model 改為 `View() tea.View`，以 `tea.NewView(content)` 或 `view.SetContent(content)` 包裝既有 render string。
5. 將 `tea.WithAltScreen()` 移除，改在相關 model view 設定 `AltScreen = true`。
6. 將 key handling 改為 `tea.KeyPressMsg`，並修正 helper function 參數型別。
7. 執行 `go mod tidy`、`go test ./...`、`make build`、`make lint`。
8. 手動 smoke test compose 與 eml TUI。

Rollback 策略：若 v2 renderer 或 Bubbles component 行為造成短期難以接受的回歸，回復 `go.mod`/`go.sum` 與 TUI API 遷移 commit，保留 `TestDrawLogo` baseline 修正作為獨立可用改進。

## Open Questions

- 是否要在同一 change 內正式修正 `TestDrawLogo`，或先以獨立小 change 處理？目前建議放在本 change 的第一個 task，因為它是升級驗證前置條件。
- 是否需要新增截圖式 golden test？目前建議先不做，除非 v2 renderer 實際造成難以手動判斷的 layout drift。
