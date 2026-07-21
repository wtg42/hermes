## MODIFIED Requirements

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

### Requirement: 快捷鍵綁定

系統 SHALL 將高頻寄送與 panel／欄位導航保留為 direct shortcuts，並將退出、清除、附件、模板與 Help 改由 `Ctrl+X` prefix command 觸發。

#### Scenario: Ctrl+S 發送郵件

- **WHEN** 使用者在撰寫畫面 normal 狀態按下 Ctrl+S
- **THEN** 系統驗證 Header 欄位，若有效則觸發發信流程（與既有設計相同）

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
