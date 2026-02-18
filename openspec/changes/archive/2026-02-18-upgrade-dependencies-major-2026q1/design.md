## Context

Hermes 是以 Go 開發的 CLI/TUI 郵件工具，核心互動面集中於 `tui/`，命令入口在 `cmd/`。目前依賴版本落後，且本次包含 `github.com/charmbracelet/bubbles` 的 major 升級，可能影響元件 API 與事件處理。

此變更需在維持既有使用流程前提下，完成依賴更新與行為驗證，並確保 `make lint`、`make test`、`make build` 可穩定通過。

## Goals / Non-Goals

**Goals:**
- 將主要直接依賴升級至目前穩定版本，建立新的依賴基線。
- 對 TUI 主要互動流程完成回歸驗證，確認升級後使用體驗不退化。
- 在必要時調整相容性程式碼與測試，維持 CLI/TUI 可預期行為。

**Non-Goals:**
- 不新增產品功能或改變既有業務邏輯。
- 不進行大規模架構重寫或 UI 風格重設計。
- 不引入與本次升級無關的新第三方框架。

## Decisions

- 決策 1：採「一次性升級 + 針對性相容修補」。
  - 理由：可一次清理累積版本落差，避免長期維護多個過渡版本。
  - 替代方案：分批升級（先 minor 再 major）風險較低，但整體週期更長。

- 決策 2：以 `go.mod` 的直接依賴為主軸，允許 indirect 依賴隨解析結果更新。
  - 理由：可避免手動鎖定大量 transitive 版本造成後續維護負擔。
  - 替代方案：手動 pin 所有 indirect 版本，短期可控但維護成本高。

- 決策 3：驗證重點放在 `tui/` 互動元件與 CLI 啟動流程。
  - 理由：本次最大不相容風險來自 TUI 生態套件版本變化。
  - 替代方案：僅執行單元測試；成本較低但無法覆蓋互動層退化。

## Risks / Trade-offs

- [Risk] `bubbles` major 變更導致 `textarea`、`textinput`、`viewport`、`filepicker` 行為差異 → Mitigation：針對相關模組補齊測試與 smoke 驗證，必要時加入相容層調整。
- [Risk] 套件升級連動 Go toolchain 或 CI 環境版本需求 → Mitigation：在 CI 與本地環境對齊 Go 版本並驗證完整 pipeline。
- [Risk] 一次性升級造成除錯面積擴大 → Mitigation：以 commit 切分「版本更新」與「相容修補」，便於定位問題與回滾。

## Migration Plan

1. 更新直接依賴版本並整理 `go.sum`。
2. 修正編譯錯誤與 API 不相容問題。
3. 執行 `make lint`、`make test`、`make build`。
4. 進行 TUI/CLI 關鍵流程 smoke 測試。
5. 若出現不可接受回歸，先回退造成問題的升級批次，再用較小批次重新推進。

## Open Questions

- 是否需要在 CI 額外加入一個最小互動 smoke job，以防未來再次升級時退化？
- 目前是否存在外部使用者依賴特定 TUI 行為（例如鍵盤快捷鍵細節）需要額外相容公告？
