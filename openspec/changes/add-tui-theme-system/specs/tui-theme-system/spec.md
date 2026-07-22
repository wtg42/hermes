## ADDED Requirements

### Requirement: 內建 Theme Registry

系統 SHALL 提供集中式 theme registry，且唯一有效的 theme 名稱為 `gruvbox` 與 `tokyo-night`；每個 theme MUST 提供完整 semantic color tokens，元件不得依賴 theme-specific token 名稱。

#### Scenario: 列出有效 Theme

- **WHEN** 系統查詢可用 theme registry
- **THEN** 結果只包含 `gruvbox` 與 `tokyo-night`，且順序穩定

#### Scenario: Theme tokens 完整

- **WHEN** 系統取得任一內建 theme
- **THEN** 該 theme 提供 Canvas、Panel、Selection、Text、Muted、Placeholder、Border、Accent、Success、Warning、Error 與 Info tokens

### Requirement: 預設與啟動 Theme 選擇

系統 SHALL 在未指定 theme 時使用 `gruvbox`，並支援以 root `--theme` flag 選擇 `gruvbox` 或 `tokyo-night`；未知名稱 MUST 在啟動 TUI 前回傳錯誤。

#### Scenario: 使用預設 Gruvbox

- **WHEN** 使用者執行 `hermes` 且未提供 `--theme`
- **THEN** Compose TUI 以 `gruvbox` 初始化

#### Scenario: 指定 Tokyo Night

- **WHEN** 使用者執行 `hermes --theme tokyo-night`
- **THEN** Compose TUI 以 `tokyo-night` 初始化

#### Scenario: 拒絕未知 Theme

- **WHEN** 使用者提供 registry 不存在的 theme 名稱
- **THEN** CLI 在啟動 TUI 前失敗，錯誤訊息列出 `gruvbox` 與 `tokyo-night`

### Requirement: Theme 套用範圍

系統 SHALL 將目前 theme 一致套用至完整 alt-screen canvas、Header／Composer／Preview panels、textinput、textarea、selection、placeholder、cursor、status bar、Filepicker、Help、安全確認與 Theme Picker overlays。

#### Scenario: Gruvbox 完整畫面

- **WHEN** Compose TUI 使用 `gruvbox` render
- **THEN** canvas 與各 panel 使用 Gruvbox backgrounds，主要文字、muted text、placeholder、focused border 與狀態色均來自 Gruvbox semantic tokens

#### Scenario: Tokyo Night 完整畫面

- **WHEN** Compose TUI 切換為 `tokyo-night` render
- **THEN** 相同元件全部改用 Tokyo Night semantic tokens，且 Compose 內容、焦點與 overlay state 保持不變

#### Scenario: Host 背景不穿透

- **WHEN** named theme 在 herdr 或其他具有自訂 pane 背景的 terminal 中 render
- **THEN** Hermes 在其 alt-screen 尺寸內以 Canvas token 填滿 panel 外空白，並以 Panel token 填滿 panel 與 textarea 空白，不混用 host pane background

#### Scenario: 邊框不使用 Terminal Default Background

- **WHEN** named theme render core panel 或 overlay border
- **THEN** 每個 border cell 均明確使用 Canvas token 作為 background，只有 border 內的內容區使用 Panel token

#### Scenario: 內建 Theme 使用校正後的視覺層級

- **WHEN** 系統解析 Gruvbox 或 Tokyo Night
- **THEN** Gruvbox Border／Accent 分別為 `#504945`／`#d79921`，Tokyo Night Border／Accent 分別為 `#3b4261`／`#7aa2f7`

### Requirement: Runtime Theme Picker

系統 SHALL 提供只包含 Gruvbox 與 Tokyo Night 的 Theme Picker overlay，支援即時預覽、確認與取消；runtime 選擇只在當次執行有效。

#### Scenario: 即時預覽 Theme

- **WHEN** Theme Picker 開啟且使用者移動到另一個 theme
- **THEN** Compose TUI 立即完整套用該 theme，且郵件草稿與焦點不變

#### Scenario: 確認 Theme

- **WHEN** 使用者在 Theme Picker 按 Enter
- **THEN** Picker 關閉並保留目前預覽的 theme

#### Scenario: 取消 Theme

- **WHEN** 使用者在 Theme Picker 預覽其他 theme 後按 Esc
- **THEN** Picker 關閉並完整恢復開啟 Picker 前的 theme

#### Scenario: 不持久化 Runtime 選擇

- **WHEN** 使用者在 Theme Picker 確認 theme 後結束並重新啟動 Hermes，且未提供 `--theme`
- **THEN** 系統仍使用預設 `gruvbox`
