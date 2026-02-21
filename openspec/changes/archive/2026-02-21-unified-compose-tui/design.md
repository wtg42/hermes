## Context

目前 hermes TUI 主流程使用 Bubbletea Elm Architecture，畫面分成兩個獨立 Model：
1. `MailFieldsModel`：7 個 textinput（From/To/Cc/Bcc/Subject/Host/Port）
2. `MailMsgModel`：textarea（內文）+ filepicker（附件）

資料透過 `viper` 全域設定系統跨層傳遞。發信邏輯在 `sendmail/use_direct_send.go` 中，與 UI 層分離。

新設計需整合上述兩個 Model 為單一 `ComposeModel`，並新增實時預覽功能。TUI 框架與外部依賴不變。

## Goals / Non-Goals

**Goals:**
- 實現單一畫面的統一撰寫 UI（左側 Header + Composer，右側 Preview）
- 支援 `Ctrl+J` / `Ctrl+K` 在 Header 和 Composer panel 間切換焦點
- 實現預覽區域即時同步 Composer 內容（純文字同步，不含 Markdown 渲染）
- 附件選取透過 Filepicker Overlay 實現
- 保持與現有發信邏輯的相容性（viper 資料傳遞、sendmail 函數）

**Non-Goals:**
- 不實作 Header/Composer panel 的展開/收合功能（`...` 和 `▽` 只是視覺裝飾）
- 不新增 Markdown 渲染（Preview 只做文字同步；glamour 為未來計劃）
- 不改變 `burst` 和 `eml` 子命令
- 不改變底層發信邏輯

## Decisions

### 1. Panel 架構與布局

**決策：** 分割畫面採用固定 50:50 左右佈局，左側包含 Header panel 和 Composer panel，右側 Preview 永遠顯示。

**理由：** 固定佈局簡化實現，無需考慮動態寬度調整。Preview 永遠顯示讓使用者隨時看到內容同步。

**替代方案考慮：**
- Panel 可摺疊（增加複雜性，用途未明確，故拒絕）
- F2 切換 Preview 顯示（浪費右側空間，不如永遠顯示）

### 2. 焦點切換鍵

**決策：** `Ctrl+J` 向下切換到 Composer，`Ctrl+K` 向上切換回 Header。

**理由：** 符合 Vim/Neovim 風格的快捷鍵習慣，易於記憶。

**替代方案考慮：**
- `Ctrl+Up/Down`（某些終端機可能無法捕捉）
- `F1` 切換（與現有 `F2 Preview` 等衝突）

### 3. Header Panel 內部導航

**決策：** `Tab` / `Shift+Tab` 在 7 個 textinput 欄位間循環移動。

**理由：** 複用 `MailFieldsModel` 既有邏輯，使用者已熟悉。

### 4. Filepicker Overlay 實現

**決策：** 狀態列 `[📎Attach]` 按鈕或 `Ctrl+A` 快捷鍵觸發 filepicker，以 overlay 形式覆蓋 Composer 區域，選擇後返回 Compose 畫面。

**理由：** 與現有 `MailMsgModel` 的 filepicker 邏輯一致，避免重複實現。Overlay 方式節省螢幕空間。

**替代方案考慮：**
- 底部直接整合 filepicker UI（浪費 Composer 空間）
- 分割視窗顯示附件列表（增加複雜性，當前未支援多附件）

### 5. Preview 的同步機制

**決策：** Preview 使用 `viewport.Model`，每次 Composer 內容更新後呼叫 `preview.SetContent(m.composer.Value())` 同步。

**理由：** `viewport` 是 charmbracelet/bubbles 既有元件，簡單可靠。純文字同步無需額外處理。

**替代方案考慮：**
- 用 textarea 代替 viewport（無法設定為只讀）
- 字符串直接拼接渲染（無法處理超過螢幕高度的內容）

### 6. 狀態列設計

**決策：** 底部狀態列顯示：`[⚡Send] [📎Attach] [F3→Quit] [Ctrl+C]` + SMTP 連線狀態（主機:埠 + TLS 狀態）

**理由：** 整合所有重要快捷鍵與連線狀態，使用者一目瞭然。

### 7. ComposeModel 的欄位結構

**決策：**
```go
type ComposeModel struct {
    // Header
    mailFields    []textinput.Model
    focusedField  int

    // Composer
    composer      textarea.Model

    // Preview
    preview       viewport.Model

    // State
    activePanel   int  // 0 = header, 1 = composer
    width, height int

    // Filepicker
    showFilePicker bool
    filepicker     filepicker.Model

    // 發信
    sending       bool
    err           error
}
```

**理由：** 清晰的欄位組織，易於區分職責；`activePanel` 明確表達焦點狀態。

## Risks / Trade-offs

| 風險 | 風險描述 | 緩解策略 |
|------|---------|---------|
| 終端機寬度不足 | 50:50 分割在寬度 < 100 字元的終端機會導致欄位截斷 | 建議最小寬度 120 字元；靠 lipgloss border 視覺提示不同 panel |
| Filepicker Overlay 的狀態管理 | Overlay 時需暫停 Composer 焦點，完成後恢復 | 新增 `showFilePicker` flag，Update 函數根據此 flag 路由按鍵事件 |
| Preview 與 Composer 的同步延遲 | 大型郵件內容（> 10K 行）可能導致 viewport 更新卡頓 | textarea 通常不會有這麼大的內容；若發生可考慮節流更新 |
| 資料傳遞相容性 | viper 資料結構需與發信邏輯相符 | 複用 `MailFieldsModel` 的欄位初始化與 viper.Set 邏輯 |

## 實作備註

- **可複用程式碼路徑：**
  - `tui/mail_field.go:40~80` - textinput 初始化
  - `tui/mail_msg_contents.go` - textarea、ctrl+h/t/e 範本、sendMailWithChannel
  - `tui/alert.go` - 發信結果提示框
  - `tui/components.go` - 按鈕樣式

- **Bubbletea 元件依賴：**
  - textinput（7 個欄位）
  - textarea（Composer）
  - viewport（Preview）
  - filepicker（Overlay）

- **按鍵路由邏輯：**
  - 全域：`ctrl+c` (Quit)、`ctrl+s` (Send)、`esc` (Clear)
  - Header：`tab/shift+tab` (欄位切換)、`ctrl+j` (切到 Composer)
  - Composer：`ctrl+k` (切到 Header)、`ctrl+h/t/e` (範本)、`ctrl+a` (Attach)
  - Filepicker Overlay：獨立的按鍵路由（複用現有 filepicker 邏輯）
