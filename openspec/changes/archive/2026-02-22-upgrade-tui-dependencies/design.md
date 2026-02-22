## Context

Hermes 使用 Charmbracelet 生態的 TUI 框架（bubbles、bubbletea、lipgloss）來構建郵件撰寫和發送的終端介面。當前版本：
- `bubbles v1.0.0`（2024 年初）
- `bubbletea v1.3.10`（2024 年中）
- `lipgloss v1.1.0`（2023 年）

最新版本已推出（v1.1+ for lipgloss、v0.32+ for bubbletea、v0.21+ for bubbles），包含性能改進、API 優化和新功能。

## Goals / Non-Goals

**Goals:**
- 升級至最新穩定版本的 bubbles、bubbletea、lipgloss
- 確保現有 TUI 功能（compose、mail_burst 等）完整運作
- 解決任何 breaking changes，適配新 API
- 運行全部測試，驗證沒有回歸

**Non-Goals:**
- 不添加新的 UI 功能（邊框標題等在後續迭代實現）
- 不進行大規模重構，只進行必要的適配
- 不改變使用者的交互方式

## Decisions

### 1. 升級順序：依賴順序安全升級
**決策**：先升級 lipgloss，再升級 bubbles，最後升級 bubbletea
**理由**：這遵循依賴樹的自下而上規則。lipgloss 是底層，bubbles 依賴 lipgloss，bubbletea 依賴 bubbles。
**替代方案**：同時升級所有 — 會增加調試難度，因為無法隔離問題來源。

### 2. 版本選擇：穩定版本優先
**決策**：選擇最新穩定版本（非 beta/alpha），除非穩定版缺乏關鍵功能
**理由**：確保生產環境穩定性。lipgloss v2.0-beta 功能新但穩定性未驗證。
**替代方案**：直接使用 v2.0-beta — 可能獲得新 Compositing API，但增加風險。

### 3. 適配策略：局部適配 + 充分測試
**決策**：更新相關代碼以適配新 API，使用現有測試驗證功能正確性
**理由**：最小化變更範圍，降低引入 bug 的風險。
**替代方案**：重寫整個 TUI 層 — 時間成本太高，風險太大。

## Risks / Trade-offs

| 風險 | 緩解方案 |
|------|---------|
| **API Breaking Changes** → 部分代碼可能無法編譯 | 檢查各套件的 CHANGELOG 和遷移指南，提前識別問題 |
| **渲染行為變化** → UI 可能看起來不同 | 充分運行測試，視覺上檢查 compose、mail_burst 等頁面 |
| **性能回歸** → 新版本在某些場景可能較慢 | 運行基準測試（如有），對比舊版本 |
| **間接依賴版本衝突** → go.mod 中的間接依賴可能產生衝突 | 使用 `go mod tidy` 和 `go get -u` 管理，必要時使用 `replace` 指令 |

## Migration Plan

1. **準備階段**
   - 檢查 bubbles、bubbletea、lipgloss 的最新版本和 CHANGELOG
   - 識別可能的 breaking changes
   - 建立獨立分支進行升級

2. **升級執行**
   - 使用 `go get` 逐個升級三個主要依賴
   - 執行 `go mod tidy` 解決間接依賴
   - 編譯代碼，解決任何編譯錯誤

3. **適配代碼**
   - 更新 `tui/compose.go`、`tui/mail_burst.go` 等中的 API 調用
   - 修復邊框、樣式、事件處理等相關代碼

4. **測試驗證**
   - 運行單元測試：`go test ./tui/...`
   - 手動測試：啟動 TUI，測試 compose、mail_burst、header、composer 等功能
   - 檢查邊框、顏色、輸入響應性

5. **合併與驗證**
   - 運行全部測試套件
   - 視覺檢查 TUI 輸出
   - 合併到 main 分支

## Open Questions

- 最新版本是否有性能影響？需要基準測試驗證嗎？
- 是否有已知的 breaking changes 需要特別關注？
- 升級後是否計劃立即實現邊框標題功能？
