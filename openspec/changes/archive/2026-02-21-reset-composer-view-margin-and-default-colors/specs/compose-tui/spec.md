## MODIFIED Requirements

### Requirement: 左側分割佈局（Header + Composer Panel）

系統 SHALL 在左側顯示兩個分割的 panel：上半部為 Header panel（7 個輸入欄位），下半部為 Composer panel（多行內文輸入）；`header composer view` 的底部間距 SHALL 使用預設值或明確為 `0`，不得額外加入非必要間距。

#### Scenario: Header Panel 顯示所有欄位

- **WHEN** 撰寫畫面載入
- **THEN** Header panel 顯示 7 個欄位：From、To、Cc、Bcc、Subject、Host、Port（順序固定）

#### Scenario: Composer Panel 允許多行編輯

- **WHEN** 焦點在 Composer panel
- **THEN** 使用者可編輯多行郵件內文，支援 Ctrl+H/Ctrl+T/Ctrl+E 快速填入 HTML/Plain Text/EML 範本

#### Scenario: Header 與 Composer 之間無額外底部間距

- **WHEN** 系統完成 Compose 畫面排版
- **THEN** `header composer view` 不會套用額外 bottom margin，呈現預設（或 `0`）的緊鄰佈局

### Requirement: 視覺設計與焦點提示

系統 SHALL 使用 lipgloss 樣式清晰標示當前焦點 panel，且輸入元件（例如 input）的前景與背景色 SHALL 使用元件預設值，不得刻意覆寫。

#### Scenario: Header Panel 邊框隨焦點改變樣式

- **WHEN** 焦點在 Header panel
- **THEN** Header panel 邊框顯示焦點樣式，Composer 邊框維持非焦點樣式

#### Scenario: Composer Panel 邊框隨焦點改變樣式

- **WHEN** 焦點在 Composer panel
- **THEN** Composer panel 邊框顯示焦點樣式，Header 邊框維持非焦點樣式

#### Scenario: 輸入元件回歸預設顏色

- **WHEN** 使用者在 Header 欄位輸入文字
- **THEN** input 顏色使用元件預設樣式，未套用自訂前景或背景色

#### Scenario: 裝飾性視覺元素

- **WHEN** 撰寫畫面初始化
- **THEN** Header 和 Composer panel 的 border title 可包含 `...` 和 `▽` 作為視覺裝飾（不實現展開/收合）
