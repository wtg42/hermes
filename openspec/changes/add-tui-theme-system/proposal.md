## Why

Compose TUI 目前將多個固定色碼直接寫在元件中，且大部分背景與 Bubble components 沿用 terminal 預設樣式；當 Hermes 執行於 herdr 等具有自身 pane theme 的環境時，文字、placeholder、selection 與背景可能失去足夠對比。現在需要建立集中式 theme system，讓整個 TUI 使用一致且可測試的語意色彩。

## What Changes

- 新增集中式 TUI theme registry 與 semantic color tokens，取代現有受版控的裸色碼。
- 只提供 `gruvbox` 與 `tokyo-night` 兩個內建 theme，預設使用 `gruvbox`。
- 將 theme 套用至完整 alt-screen canvas、Header／Composer／Preview panels、textinput、textarea、selection、placeholder、cursor、status bar 與 overlays。
- 新增 `--theme` 啟動參數，接受 `gruvbox` 或 `tokyo-night`，無效名稱須明確失敗。
- 在既有 `Ctrl+X` prefix registry 新增 `p` Palette command，顯示可即時預覽的 theme picker；Enter 套用，Esc 取消並恢復原 theme。
- Theme runtime 切換只影響當次執行，不建立新的持久化 config system。
- 移除 Compose TUI 與共用元件中的固定 `#DC851C`、`#383838`、ANSI `214`、`196`、`240` 等顏色來源。

## Capabilities

### New Capabilities

- `tui-theme-system`: 定義 theme registry、Gruvbox／Tokyo Night palettes、預設 theme、CLI 選擇、runtime theme picker 與完整元件套用行為。

### Modified Capabilities

- `compose-tui`: 將原本使用預設元件顏色與固定焦點色的視覺規格改為由目前 theme 統一控制 canvas、panels、inputs、status 與 overlays。
- `tui-command-prefix`: 在 root prefix registry 加入 `p` Palette command，並定義 theme picker 的按鍵隔離與取消行為。

## Impact

- 主要影響 `tui/compose.go`、`tui/components.go`、`tui/command_prefix.go`、TUI 初始化流程及相關 Bubble Tea v2 tests。
- CLI root command 會新增 `--theme` flag，但不改變 SMTP、Burst、EML 或郵件資料模型。
- 不新增外部 dependency；沿用 Lip Gloss v2 與 Bubbles v2 的 style APIs。
- Runtime theme picker 不寫入使用者檔案，避免本次變更擴張為完整設定管理系統。
