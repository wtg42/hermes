## Context

Hermes Compose TUI 目前在 `tui/components.go` 與 `tui/compose.go` 直接建立 Lip Gloss styles，並混用 `#DC851C`、`#383838` 與 ANSI `214`、`196`、`240`。Header、Composer、Preview 與 Bubbles inputs 沒有共同 palette，也未替完整 alt-screen canvas 指定背景。當程式執行在 herdr、tmux 或具有自訂 terminal theme 的環境中，空白區域會透出 host 背景，元件預設文字與 Hermes 固定 muted color 也可能形成低對比組合。

本變更橫跨 CLI 啟動、Compose model、prefix command、Bubble components 與 View rendering，因此需要先定義 theme ownership、runtime preview rollback 與 style 套用邊界。

## Goals / Non-Goals

**Goals:**

- 以 semantic tokens 集中管理所有 Compose TUI 色彩。
- 只提供 Gruvbox 與 Tokyo Night，並以 Gruvbox 作為無設定時的預設。
- 讓完整 canvas、panels、inputs、selection、cursor、status 與 overlays 使用同一個 theme。
- 支援 `--theme` 啟動選擇與 `Ctrl+X`、`p` runtime Theme Picker。
- Theme Picker 能即時預覽，Enter 保留選擇，Esc 恢復開啟前 theme。
- 以 Bubble Tea v2 state tests 與固定 color profile View tests 驗證行為。

**Non-Goals:**

- 不提供 Gruvbox／Tokyo Night 以外的 theme。
- 不支援使用者自訂 token、匯入 theme 或 theme plugin。
- 不新增 config file、寫入 `$HOME` 或持久保存 runtime 選擇。
- 不實作 host terminal light/dark auto detection 或自動切換。
- 不改變 AlertModel 以外的舊 TUI 流程、SMTP、Burst 或 EML 寄送行為。

## Decisions

### 1. Theme registry 只保存具名 semantic palette

新增獨立 theme module，定義 `Theme` 與固定 registry。Theme 至少包含 Canvas、Panel、Selection、Text、Muted、Placeholder、Border、Accent、Success、Warning、Error 與 Info tokens。元件只使用 token，不直接引用 hex 或 ANSI color。

採 semantic tokens 而不是將 Gruvbox 的 `bg0`、`fg1` 等原始名稱傳入 View，讓相同元件能在 Tokyo Night 使用不同 palette 而不改 rendering code。Registry 僅含兩個 immutable built-ins，避免過早設計外部 theme schema。

### 2. Gruvbox 是預設，CLI resolver 對未知名稱 fail closed

未提供 theme 時解析為 `gruvbox`。`--theme` 接受 `gruvbox` 或 `tokyo-night`；未知值在啟動 TUI 前回傳包含有效名稱的錯誤，不默默 fallback。這讓 typo 可見，也讓測試與畫面可重現。

替代方案是接受任意字串後 fallback，但這會讓使用者誤以為選擇已生效，因此不採用。

### 3. ComposeModel 擁有目前 theme，style 套用集中在單一方法

ComposeModel 保存目前 theme name／value，初始化與 runtime 切換都呼叫同一個 `applyTheme`。該方法同步更新 Header textinputs、Composer textarea、Preview viewport、Filepicker 及 model-level rendering styles，避免切換後只有邊框或背景改色。

既有 `InitialComposeModel(mailer)` 保留為 Gruvbox default 入口，另提供可注入 theme 的初始化路徑給 CLI 與 tests，避免所有呼叫端一次承受 signature breaking change。

### 4. Named theme 填滿完整 alt-screen canvas

View 最外層以目前 theme 的 Canvas token render terminal 尺寸內的所有空白；Panel token 則只用於 Header、Composer、Preview 與 overlays。如此 Hermes 不會意外混入 herdr pane background，panel 層級仍可辨識。

只替 panel 設背景會讓 margins、status row 與窄視窗補白透出 host 顏色，因此不採用。

### 5. Theme Picker 是 overlay state，preview 可 rollback

Root prefix registry 新增 `p` Palette。觸發後建立 picker state，保存 original theme 與目前 index。上下方向鍵或 `j`／`k` 改變 index並立即呼叫 `applyTheme`；Enter 關閉 picker並保留目前 theme；Esc 關閉 picker並重新套用 original theme。

Theme Picker 開啟時優先接收所有 key，避免按鍵穿透至 prefix、Header 或 Composer。Picker 不寫入檔案，重啟後仍依 CLI flag 或 Gruvbox default。

### 6. Bubble component styles 由 Theme 明確設定

Textinput 與 textarea 使用 Bubbles v2 `Styles()`／`SetStyles()` 更新 focused、blurred、placeholder、prompt、cursor、line number 與 selection 相關樣式；Viewport、Filepicker 與 overlays 使用對應 Lip Gloss style。這解決目前只改外框、元件內容仍沿用不相容 terminal foreground 的問題。

### 7. 測試將 raw color 與 theme state 視為 contract

先新增 registry、default／invalid resolution、component style、picker preview／confirm／cancel 與 View tests，再進行實作。測試固定 window size、environment 與 color profile，避免依賴開發者 terminal。另加入 source guard，確保受影響 TUI rendering 不再新增散落的裸色碼。

### 8. Textarea 以 ANSI-aware cell pass 補齊透明背景

Bubbles v2 textarea 的內部 viewport 會以沒有 background style 的空白 cell 補滿空行；內層 ANSI reset 也會中斷外層 Lip Gloss Panel background，讓 host terminal 背景穿透。Composer render 後使用既有 Ultraviolet parser 分解 ANSI cells，只為沒有 background 的 cell 補上目前 Theme 的 Panel token，再重新 render。已明確設定的 Selection、文字、cursor line 與 line number styles 必須保留。

### 9. Panel 使用直角低調邊框與左側 focus rail

Core panels 使用 Normal border。Inactive panel 四邊皆使用 Border token；focused Header 或 Composer 僅左邊使用 Accent，其餘三邊維持 Border。Preview 不可顯示 focus rail。Overlays 同樣改用直角 border，但依其語意使用 Accent、Warning 或 Error。Gruvbox Border／Accent 校正為 `#504945`／`#d79921`，Tokyo Night Border 校正為 `#3b4261` 並保留 `#7aa2f7` Accent。

### 10. Border background 明確屬於 Canvas

Lip Gloss border glyph 若沒有 background 會回落至 terminal default，導致水平 border 在 herdr 中呈現黑色橫條。所有 core panels 與 overlays 透過同一個 bordered-panel helper，將內容 background 設為 Panel、四側及 corners 的 border background 設為 Canvas。Focused panel 只改變左側 border foreground，不改變任何一側的 Canvas background。

## Risks / Trade-offs

- [Risk] 為完整 canvas 填色可能讓 NoColor terminal 產生多餘 ANSI sequence。→ 透過 Bubble Tea color profile 與 NoColor View tests 驗證降級行為。
- [Risk] Bubbles v2 各元件 style 欄位不同，runtime 切換可能遺漏子樣式。→ 集中 `applyTheme` 並以 component-level tests 驗證 placeholder、selection 與 focused／blurred state。
- [Risk] Theme Picker 即時 preview 會改變 model，取消時若只改 theme name 可能留下舊 style。→ 保存 original theme 並以同一個 `applyTheme` 完整 rollback。
- [Risk] `p` 未來可能與其他 prefix command 衝突。→ command registry 作為 key 與 HUD 的唯一來源，由 registry tests 防止重複 key。
- [Risk] ANSI-aware cell pass 可能覆蓋 textarea 既有 selection。→ 只填補 background 為空的 cell，並逐 cell 測試 Selection 保留。
- [Risk] Inline 建立 border style 容易遺漏某一側 background。→ 集中 bordered-panel helper，並對四側與 corners 做 cell-level regression tests。
- [Trade-off] 本次不持久化 runtime 選擇。→ 使用者可用 `--theme` 重現選擇；持久 config 另開 change。

## Migration Plan

1. 新增 theme registry 與 tests，保留既有 rendering 行為直到 component migration 完成。
2. 將 Compose、共用 components 與 overlays 的固定色碼改為 semantic tokens。
3. 將 Gruvbox 設為 default 並新增 `--theme` resolver。
4. 接上 prefix Palette command 與 Theme Picker overlay。
5. 執行 View、unit、race、integration 與 OpenSpec validation。

若需回滾，可移除 picker 與 CLI flag，讓 `InitialComposeModel` 固定建立 Gruvbox theme；郵件與資料格式不涉及 migration。

## Open Questions

無。Theme 持久化、自訂 palette 與更多內建 themes 明確留待後續 change。
