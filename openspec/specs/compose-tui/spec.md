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

系統 SHALL 在左側顯示兩個分割的 panel：上半部為 Header panel（7 個輸入欄位），下半部為 Composer panel（多行內文輸入）；`header composer view` 的底部間距 SHALL 使用預設值或明確為 `0`，不得額外加入非必要間距。

#### Scenario: Header Panel 顯示所有欄位

- **WHEN** 撰寫畫面載入
- **THEN** Header panel 顯示 7 個欄位：From、To、Cc、Bcc、Subject、Host、Port（順序固定）

#### Scenario: Composer Panel 允許多行編輯

- **WHEN** 焦點在 Composer panel
- **THEN** 使用者可編輯多行郵件內文，並可透過 `Ctrl+X`、`t` 進入 template prefix 後選擇 HTML、Plain Text 或 EML 範本

#### Scenario: Header 與 Composer 之間無額外底部間距

- **WHEN** 系統完成 Compose 畫面排版
- **THEN** `header composer view` 不會套用額外 bottom margin，呈現預設（或 `0`）的緊鄰佈局

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

系統 SHALL 在底部顯示狀態列，包含當下狀態可用的快捷鍵提示、prefix command 入口與 SMTP 連線狀態。

#### Scenario: Normal 狀態顯示精簡快捷鍵

- **WHEN** 撰寫畫面處於 normal 狀態且未在寄送
- **THEN** 底部狀態列顯示 `[Ctrl+S] Send`、`[Ctrl+X] Commands` 以及目前 panel 的導航提示

#### Scenario: Prefix 狀態顯示 Command HUD

- **WHEN** 使用者按下 `Ctrl+X` 進入 prefix command 模式
- **THEN** 底部狀態列暫時改為顯示 root Command HUD，而不是 normal 狀態快捷鍵

#### Scenario: 狀態列顯示 SMTP 連線狀態

- **WHEN** 使用者填入 Host 和 Port 後
- **THEN** 底部狀態列顯示「Connected to smtp.example.com:587 • TLS active」（或對應的主機與埠）

#### Scenario: 狀態列實時更新連線資訊

- **WHEN** 使用者修改 Host 或 Port 欄位
- **THEN** 狀態列的連線資訊即時更新

---

### Requirement: Filepicker Overlay 附件選取

系統 SHALL 透過 Overlay 方式實現附件選取。使用者執行 prefix Attach command 時，filepicker 以全屏 overlay 覆蓋 Composer 區域。

#### Scenario: 觸發 Filepicker Overlay

- **WHEN** 焦點在 Header 或 Composer，且使用者依序按下 `Ctrl+X`、`a`
- **THEN** Filepicker Overlay 出現，使用者可選擇附件檔案

#### Scenario: 選擇附件後返回撰寫畫面

- **WHEN** 使用者在 Filepicker 中選擇檔案並確認
- **THEN** Overlay 關閉，焦點返回撰寫畫面（保持之前的 panel），被選檔案路徑被記錄

#### Scenario: 取消 Filepicker Overlay

- **WHEN** 使用者在 Filepicker 中按 Esc
- **THEN** Overlay 關閉，焦點返回撰寫畫面，不選擇任何檔案，且 Compose 內容保持不變

---

### Requirement: 快捷鍵綁定

系統 SHALL 將高頻寄送與 panel／欄位導航保留為 direct shortcuts，並將退出、清除、附件、模板與 Help 改由 `Ctrl+X` prefix command 觸發。

#### Scenario: Ctrl+S 發送郵件

- **WHEN** 使用者在撰寫畫面 normal 狀態按下 Ctrl+S
- **THEN** 系統驗證 Header 欄位，若有效則觸發發信流程（與舊設計相同）

#### Scenario: Prefix Quit 結束程式

- **WHEN** Compose 為空且使用者依序按下 `Ctrl+X`、`q`
- **THEN** 系統終止 TUI

#### Scenario: Ctrl+C 不再立即結束

- **WHEN** 使用者在 Compose TUI 按下 Ctrl+C
- **THEN** 系統不得將其視為立即退出指令，退出 SHALL 經由 prefix Quit flow

#### Scenario: Esc 不再清空或退出

- **WHEN** 使用者在 normal Compose 狀態按下 Esc
- **THEN** 系統不得清空 Header、Body 或附件，也不得終止 TUI

#### Scenario: Prefix Clear 清空欄位

- **WHEN** Compose 含有內容，且使用者依序按下 `Ctrl+X`、`c`、`c`
- **THEN** 系統清空所有 Header 欄位、Composer、Preview 與附件

#### Scenario: Prefix Template 套用模板

- **WHEN** 焦點在 Composer，且使用者透過 `Ctrl+X`、`t` 進入 template prefix 後選擇模板
- **THEN** 對應的 HTML、Plain Text 或 EML 範本被填入 Composer，Preview 同步更新

#### Scenario: 舊附件與模板 direct shortcuts 不再觸發應用程式操作

- **WHEN** 使用者按下 `Ctrl+A`、`Ctrl+H`、`Ctrl+T` 或 `Ctrl+E`
- **THEN** 系統不得以這些按鍵直接開啟 Filepicker 或套用 Compose 模板

#### Scenario: Panel 與欄位導航維持不變

- **WHEN** 使用者按下 `Ctrl+J`、`Ctrl+K`、Tab 或 Shift+Tab
- **THEN** 系統維持既有 Header／Composer panel 與 Header 欄位導航行為

---

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

---

### Requirement: Compose TUI 安全確認狀態

Compose TUI SHALL 在既有撰寫畫面內提供白名單外寄信確認狀態，清楚顯示彙整原因、精確 `SEND` 指示、確認輸入與取消方式。確認畫面 MUST 保持在同一個 Bubble Tea program 與 ComposeModel 狀態中。

#### Scenario: 顯示安全警告內容

- **WHEN** ComposeModel 進入 pending confirmation state
- **THEN** View 顯示每個越界原因、`Type SEND to confirm` 指示與 Esc 取消提示

#### Scenario: 確認輸入取得焦點

- **WHEN** 安全確認畫面首次顯示
- **THEN** 鍵盤輸入由確認欄位處理，不會傳入 Header、Composer、Preview 或 Filepicker

#### Scenario: 取消後恢復撰寫畫面

- **WHEN** 使用者按 Esc 取消安全確認
- **THEN** View 返回原 Compose 畫面，所有 Header、Body 與附件選擇保持不變

---

### Requirement: Compose TUI 寄信路由依安全評估分流

ComposeModel SHALL 在 `Ctrl+S` 時先執行共用單封安全評估，再決定直接寄送、顯示格式錯誤或進入確認狀態。

#### Scenario: 安全郵件直接寄送

- **WHEN** 使用者對全部位於安全邊界且格式有效的郵件按下 `Ctrl+S`
- **THEN** TUI 不顯示確認畫面，沿用既有 sending state 與非同步 Mailer command

#### Scenario: 格式錯誤不進入安全確認

- **WHEN** 使用者對包含無效地址或 port 的郵件按下 `Ctrl+S`
- **THEN** TUI 顯示具體格式錯誤、不進入確認狀態且不呼叫 Mailer

#### Scenario: 白名單外郵件延後寄送

- **WHEN** 使用者對格式有效但含有越界原因的郵件按下 `Ctrl+S`
- **THEN** TUI 進入 pending confirmation state，直到精確確認前都不建立寄信 command

#### Scenario: 正確確認後只寄送一次

- **WHEN** 使用者輸入精確 `SEND` 且快照驗證通過
- **THEN** ComposeModel 清除確認狀態、設為 sending 並建立一個 Mailer command

---

### Requirement: 安全確認期間快捷鍵隔離

ComposeModel MUST 在 pending confirmation state 優先處理確認按鍵，且 SHALL NOT 讓一般全域或 panel 快捷鍵在背景生效。

#### Scenario: 確認期間再次按 Ctrl+S

- **WHEN** 使用者在確認畫面再次按下 `Ctrl+S`
- **THEN** 系統不建立寄信 command，也不疊加第二個確認狀態

#### Scenario: 確認期間按附件或模板快捷鍵

- **WHEN** 使用者在確認畫面按下 `Ctrl+A`、`Ctrl+H`、`Ctrl+T` 或 `Ctrl+E`
- **THEN** 系統不開啟 Filepicker、不替換 Composer 內容，並維持相同 pending confirmation
