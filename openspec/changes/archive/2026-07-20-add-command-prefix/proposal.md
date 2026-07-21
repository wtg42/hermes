## Why

Compose TUI 目前將寄送、附件、清除、模板與退出等功能分散在多組全域快捷鍵，狀態列也會隨功能增加而愈來愈擁擠。加入可顯示指令提示的 prefix command 模式，可提供一致且可擴充的鍵盤操作入口，並降低誤觸退出或清除內容的風險。

## What Changes

- 新增以 `Ctrl+X` 啟動的 prefix command 模式，啟動後在狀態列顯示當下可用指令。
- 新增共用 command registry／dispatcher，集中定義指令識別、按鍵、標籤、啟用條件與行為。
- 將 Quit、Clear、Attachment、Template 與 Help 放入 prefix command namespace。
- 保留 `Ctrl+S` 寄送、`Ctrl+J`／`Ctrl+K` panel 導航及 Tab 欄位導航等高頻直接操作。
- 將 `Esc` 統一為取消 prefix、關閉 overlay 或返回上一層，不再直接清除 Compose 或退出。
- 對有內容的 Compose 執行 Clear 或 Quit 時提供確認流程，避免無意遺失輸入。
- **BREAKING**：Compose TUI 的主要退出操作由 `Ctrl+C`／連按 `Esc` 改為 `Ctrl+X` 後按 `q`。
- **BREAKING**：附件選取與模板套用改由 prefix command 入口觸發，不再以既有的 `Ctrl+A`、`Ctrl+H`、`Ctrl+T`、`Ctrl+E` 作為主要操作。

## Capabilities

### New Capabilities

- `tui-command-prefix`: 定義 prefix command 狀態、Command HUD、指令分派、取消與確認互動。

### Modified Capabilities

- `compose-tui`: 更新 Compose 的狀態列、退出、清除、附件、模板與 Esc 快捷鍵行為，使其整合 prefix command。

## Impact

- 主要影響 `tui/compose.go`、TUI 元件與快捷鍵相關測試。
- 將新增上層 command routing 狀態與可測試的指令定義，作為未來 History 等畫面導航的基礎。
- 不改變 SMTP 發信 API、EML 子命令或 Burst mode 行為。
- 不需新增外部 dependency。
