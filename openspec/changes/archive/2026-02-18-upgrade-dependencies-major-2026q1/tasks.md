## 1. 依賴升級與基線整理

- [x] 1.1 盤點並更新本次範圍內的直接依賴版本（含 `bubbles` major 升級）
- [x] 1.2 執行 `go mod tidy` 與必要的 module 解析，更新 `go.sum` 並確認無異常依賴衝突
- [x] 1.3 檢查 Go toolchain 與 CI 設定是否滿足升級後依賴需求

## 2. 相容性修補與測試補強

- [x] 2.1 修正因依賴升級導致的編譯錯誤與 API 不相容問題（優先處理 `tui/` 元件）
- [x] 2.2 補齊或調整受影響測試，涵蓋 TUI 核心互動與 CLI 關鍵路徑
- [x] 2.3 針對高風險模組（`textarea`、`textinput`、`viewport`、`filepicker`）完成回歸確認

## 3. 驗證、風險控管與交付

- [x] 3.1 執行 `make lint`、`make test`、`make build` 並記錄結果
- [x] 3.2 執行 CLI/TUI smoke 測試並確認主要操作流程可用
- [x] 3.3 若出現重大回歸，依回復策略回退問題升級批次並重新驗證（本次未觸發）
