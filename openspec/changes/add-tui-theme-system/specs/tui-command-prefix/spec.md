## MODIFIED Requirements

### Requirement: Command HUD

系統 SHALL 在正常狀態列顯示 `[Ctrl+X] Commands`，並在 prefix command 模式以 Command HUD 顯示目前可用的下一鍵指令與取消方式。

#### Scenario: 正常狀態顯示 Commands 入口

- **WHEN** Compose TUI 處於 normal 狀態且未在寄送
- **THEN** 狀態列顯示 `[Ctrl+X] Commands` 以及目前 panel 的高頻操作提示

#### Scenario: Root prefix 顯示可用指令

- **WHEN** 使用者按下 `Ctrl+X`
- **THEN** 狀態列顯示 `COMMAND`，並列出 Attach、Template、Palette、Clear、Quit、Help 與 Esc Cancel

#### Scenario: Template prefix 顯示模板選項

- **WHEN** 使用者在 root prefix 狀態按下 `t`
- **THEN** 狀態列顯示 HTML、Plain Text、EML template 與 Esc Back 的下一鍵選項

### Requirement: Prefix 指令分派

系統 SHALL 透過集中定義的 command registry 將 prefix key 分派為 Attach、Template、Palette、Clear、Quit 或 Help semantic command，並確保顯示 label 與實際 key binding 來自相同定義。

#### Scenario: 執行附件指令

- **WHEN** 使用者依序按下 `Ctrl+X`、`a`
- **THEN** 系統執行 Attach command 並開啟現有 Filepicker Overlay

#### Scenario: 進入模板指令

- **WHEN** 使用者依序按下 `Ctrl+X`、`t`
- **THEN** 系統進入 template prefix，等待使用者選擇 HTML、Plain Text 或 EML template

#### Scenario: 開啟 Theme Picker

- **WHEN** 使用者依序按下 `Ctrl+X`、`p`
- **THEN** 系統執行 Palette command 並開啟 Theme Picker Overlay

#### Scenario: 顯示 Help

- **WHEN** 使用者依序按下 `Ctrl+X`、`?`
- **THEN** 系統顯示包含 direct shortcuts 與 prefix commands 的 Help overlay，且內容包含 Palette command

### Requirement: Overlay 與 prefix 的按鍵隔離

系統 SHALL 優先將按鍵交給目前開啟的 overlay 或確認狀態，避免 prefix command 或背景 Compose 操作穿透。

#### Scenario: Filepicker 中按 Esc

- **WHEN** Filepicker Overlay 已開啟且使用者按下 Esc
- **THEN** 系統只關閉 Filepicker，返回原 Compose 焦點，不清除內容也不退出

#### Scenario: Theme Picker 中輸入導航鍵

- **WHEN** Theme Picker Overlay 已開啟且使用者按下方向鍵、`j`、`k`、Enter 或 Esc
- **THEN** 該按鍵只由 Theme Picker 處理，不得觸發 prefix、Header、Composer、寄送或退出操作

#### Scenario: Prefix key 不寫入輸入欄位

- **WHEN** 使用者在任一 prefix 或確認狀態輸入 command key
- **THEN** 該按鍵只由 command dispatcher 處理，不得傳遞給 Header textinput 或 Composer textarea
