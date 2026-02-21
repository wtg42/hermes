## Why

目前 Composer 畫面的 `header composer view` 底部間距與部分元件（例如 input）的色彩樣式有刻意覆寫，造成版面間距與視覺表現偏離預設行為，增加維護成本，也讓整體 UI 一致性下降。需要將這些樣式回歸預設（或等效為 `0` 間距），以降低不必要的客製化風險。

## What Changes

- 將 `header composer view` 的 bottom margin 調整為預設值，或明確設為 `0` 以達成等效行為。
- 移除非必要的元件色彩覆寫（例如 input 顏色），改用框架/元件庫的預設樣式。
- 盤點並清理同類型的硬編碼色彩設定，確保未再刻意指定自訂顏色。

## Capabilities

### New Capabilities

- （無）

### Modified Capabilities

- `compose-tui`: 調整 Composer 介面樣式需求，要求 header 間距與輸入元件色彩回歸預設，避免不必要的樣式覆寫。

## Impact

- 受影響範圍：`tui/` 內 Composer 相關視圖與樣式定義。
- 對外 API/CLI 行為無破壞性變更，主要是 UI 樣式一致性與可維護性提升。
- 測試可能需更新快照或 UI 行為驗證（若既有測試依賴具體樣式值）。
