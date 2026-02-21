# compose-tui Specification

## Purpose
TBD - created by archiving change unified-compose-tui. Update Purpose after archive.
## Requirements
### Requirement: 統一撰寫畫面

系統 SHALL 提供單一撰寫畫面（ComposeModel），整合郵件 Header 欄位、Composer 多行內文、右側 Preview 實時同步，並支援附件選取。

#### Scenario: 使用者啟動 hermes 進入撰寫畫面

- **WHEN** 使用者執行 `hermes` 命令（無子命令、無旗標）
- **THEN** 系統啟動 TUI，顯示統一撰寫畫面（而非舊的兩步驟流程）

#### Scenario: 首次進入時的焦點

- **WHEN** 撰寫畫面初始化完成
- **THEN** 焦點預設在 Header panel 的第一個欄位（From）

---

### Requirement: 左側分割佈局（Header + Composer Panel）

系統 SHALL 在左側顯示兩個分割的 panel：上半部為 Header panel（7 個輸入欄位），下半部為 Composer panel（多行內文輸入）。

#### Scenario: Header Panel 顯示所有欄位

- **WHEN** 撰寫畫面載入
- **THEN** Header panel 顯示 7 個欄位：From、To、Cc、Bcc、Subject、Host、Port（順序固定）

#### Scenario: Composer Panel 允許多行編輯

- **WHEN** 焦點在 Composer panel
- **THEN** 使用者可編輯多行郵件內文，支援 Ctrl+H/Ctrl+T/Ctrl+E 快速填入 HTML/Plain Text/EML 範本

---

### Requirement: 右側 Preview Panel

系統 SHALL 在右側顯示 Preview panel，以只讀的 viewport 即時同步 Composer 內容（純文字，不含 Markdown 渲染）。

#### Scenario: Preview 同步 Composer 內容

- **WHEN** 使用者在 Composer 輸入或編輯文字
- **THEN** Preview panel 即時更新，顯示完全相同的文字內容

#### Scenario: Preview 適應終端機視窗高度

- **WHEN** 內文超過 Preview panel 的可顯示行數
- **THEN** Preview 顯示垂直捲軸，使用者可上下捲動預覽內容

#### Scenario: Preview 在焦點切換時保持顯示

- **WHEN** 焦點在 Header 或 Composer
- **THEN** Preview panel 始終顯示（不隱藏、不收合）

---

### Requirement: Ctrl+J / Ctrl+K 焦點切換

系統 SHALL 支援 `Ctrl+J` 和 `Ctrl+K` 快捷鍵在 Header 和 Composer panel 間切換焦點。

#### Scenario: Ctrl+J 從 Header 切換到 Composer

- **WHEN** 焦點在 Header panel（任何欄位）且按下 Ctrl+J
- **THEN** 焦點切換到 Composer panel 的 textarea，textarea 獲得焦點並可立即編輯

#### Scenario: Ctrl+K 從 Composer 切換回 Header

- **WHEN** 焦點在 Composer panel 且按下 Ctrl+K
- **THEN** 焦點切換到 Header panel，回到上次在 Header 中焦點的欄位（或第一個欄位）

#### Scenario: 在 Header panel 內循環導航

- **WHEN** 焦點在 Header panel 且按下 Tab
- **THEN** 焦點移動到下一個欄位；若已在最後一個欄位，Tab 不切換到 Composer（保持在 Header 內）

#### Scenario: Shift+Tab 在 Header 內向後導航

- **WHEN** 焦點在 Header panel 且按下 Shift+Tab
- **THEN** 焦點移動到前一個欄位；若已在第一個欄位，Shift+Tab 循環到最後一個欄位

---

### Requirement: 底部狀態列

系統 SHALL 在底部顯示狀態列，包含快捷鍵提示與 SMTP 連線狀態。

#### Scenario: 狀態列顯示所有快捷鍵

- **WHEN** 撰寫畫面顯示
- **THEN** 底部狀態列顯示：`[⚡Send] [📎Attach] [F3→Quit] [Ctrl+C]`

#### Scenario: 狀態列顯示 SMTP 連線狀態

- **WHEN** 使用者填入 Host 和 Port 後
- **THEN** 底部狀態列右側顯示「Connected to smtp.example.com:587 • TLS active」（或對應的主機與埠）

#### Scenario: 狀態列實時更新連線資訊

- **WHEN** 使用者修改 Host 或 Port 欄位
- **THEN** 狀態列的連線資訊即時更新

---

### Requirement: Filepicker Overlay 附件選取

系統 SHALL 透過 Overlay 方式實現附件選取。按 `[📎Attach]` 按鈕或 `Ctrl+A` 快捷鍵時，filepicker 以全屏 overlay 覆蓋 Composer 區域。

#### Scenario: 觸發 Filepicker Overlay

- **WHEN** 焦點在 Header 或 Composer 且按下 Ctrl+A（或點擊狀態列的 `[📎Attach]`）
- **THEN** Filepicker Overlay 出現，使用者可選擇附件檔案

#### Scenario: 選擇附件後返回撰寫畫面

- **WHEN** 使用者在 Filepicker 中選擇檔案並確認
- **THEN** Overlay 關閉，焦點返回撰寫畫面（保持之前的 panel），被選檔案路徑被記錄

#### Scenario: 取消 Filepicker Overlay

- **WHEN** 使用者在 Filepicker 中按 Esc 或點擊取消
- **THEN** Overlay 關閉，焦點返回撰寫畫面，不選擇任何檔案

---

### Requirement: 快捷鍵綁定

系統 SHALL 支援以下全域快捷鍵：

#### Scenario: Ctrl+S 發送郵件

- **WHEN** 使用者在撰寫畫面按下 Ctrl+S
- **THEN** 系統驗證 Header 欄位，若有效則觸發發信流程（與舊設計相同）

#### Scenario: Ctrl+C 強制結束

- **WHEN** 使用者按下 Ctrl+C
- **THEN** 系統立即終止 TUI，不保存任何內容

#### Scenario: Esc 清空欄位

- **WHEN** 使用者按下 Esc（第一次）
- **THEN** 所有 Header 欄位與 Composer 內容被清空；連按兩次 Esc 則直接退出

#### Scenario: 模板快捷鍵在 Composer 中生效

- **WHEN** 焦點在 Composer 且按下 Ctrl+H、Ctrl+T 或 Ctrl+E
- **THEN** 對應的 HTML/Plain Text/EML 範本被填入 Composer（複用舊設計邏輯）

---

### Requirement: 視覺設計與焦點提示

系統 SHALL 使用 lipgloss 樣式清晰標示當前焦點 panel。

#### Scenario: Header Panel 邊框隨焦點改變顏色

- **WHEN** 焦點在 Header panel
- **THEN** Header panel 邊框顯示為橘色（`#DC851C`），Composer 邊框為普通灰色

#### Scenario: Composer Panel 邊框隨焦點改變顏色

- **WHEN** 焦點在 Composer panel
- **THEN** Composer panel 邊框顯示為橘色，Header 邊框為普通灰色

#### Scenario: 裝飾性視覺元素

- **WHEN** 撰寫畫面初始化
- **THEN** Header 和 Composer panel 的 border title 可包含 `...` 和 `▽` 作為視覺裝飾（不實現展開/收合）

---

### Requirement: 資料流與發信整合

系統 SHALL 保持與現有發信邏輯相容。所有 Header 欄位與 Composer 內容透過 viper 全域設定系統傳遞，使用現有的 `sendmail.SendMailWithMultipart` 函數發信。

#### Scenario: 發信前驗證欄位

- **WHEN** 使用者按 Ctrl+S
- **THEN** 系統驗證 To、Cc、Bcc 欄位的郵件地址有效性；若無效則顯示錯誤提示（複用舊設計）

#### Scenario: 發信結果提示

- **WHEN** 發信完成（成功或失敗）
- **THEN** 系統使用 AlertModel 顯示結果提示框（複用舊設計）

#### Scenario: 發信後返回撰寫畫面

- **WHEN** 使用者在結果提示框中按 Esc
- **THEN** AlertModel 關閉，焦點返回撰寫畫面（Header panel 預設焦點）

