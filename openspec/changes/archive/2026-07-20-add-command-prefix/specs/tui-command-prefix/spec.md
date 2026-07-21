## ADDED Requirements

### Requirement: Prefix command 模式

系統 SHALL 在 Compose TUI 的 Header 或 Composer panel 中，透過 `Ctrl+X` 進入 prefix command 模式，並攔截下一個按鍵以分派指令，而不得將該按鍵寫入目前輸入元件。

#### Scenario: 從 Header 進入 prefix command

- **WHEN** 使用者在 Header panel 按下 `Ctrl+X`
- **THEN** 系統進入 root prefix 狀態，保留目前焦點與所有 Compose 內容

#### Scenario: 從 Composer 進入 prefix command

- **WHEN** 使用者在 Composer panel 按下 `Ctrl+X`
- **THEN** 系統進入 root prefix 狀態，且下一個 command key 不會被輸入 Body

#### Scenario: Prefix 不自動逾時

- **WHEN** 系統已進入 prefix command 模式且尚未收到下一個按鍵
- **THEN** 系統 SHALL 保持 prefix command 模式，直到指令執行、取消或收到未知 key

### Requirement: Command HUD

系統 SHALL 在正常狀態列顯示 `[Ctrl+X] Commands`，並在 prefix command 模式以 Command HUD 顯示目前可用的下一鍵指令與取消方式。

#### Scenario: 正常狀態顯示 Commands 入口

- **WHEN** Compose TUI 處於 normal 狀態且未在寄送
- **THEN** 狀態列顯示 `[Ctrl+X] Commands` 以及目前 panel 的高頻操作提示

#### Scenario: Root prefix 顯示可用指令

- **WHEN** 使用者按下 `Ctrl+X`
- **THEN** 狀態列顯示 `COMMAND`，並列出 Attach、Template、Clear、Quit、Help 與 Esc Cancel

#### Scenario: Template prefix 顯示模板選項

- **WHEN** 使用者在 root prefix 狀態按下 `t`
- **THEN** 狀態列顯示 HTML、Plain Text、EML template 與 Esc Back 的下一鍵選項

### Requirement: Prefix 取消與未知指令

系統 SHALL 允許使用者以 Esc 安全取消 prefix command，且未知 key 不得改變 Compose 內容或執行其他指令。

#### Scenario: Esc 取消 root prefix

- **WHEN** 使用者在 root prefix 狀態按下 Esc
- **THEN** 系統返回 normal 狀態，恢復原焦點且保留所有 Compose 內容

#### Scenario: Esc 返回 template prefix 上一層

- **WHEN** 使用者在 template prefix 狀態按下 Esc
- **THEN** 系統返回 root prefix 並顯示 root Command HUD

#### Scenario: 輸入未知指令

- **WHEN** 使用者在 root prefix 狀態按下 registry 未定義的 key
- **THEN** 系統不執行任何 command、返回 normal 狀態，並顯示 unknown command 訊息

### Requirement: Prefix 指令分派

系統 SHALL 透過集中定義的 command registry 將 prefix key 分派為 Attach、Template、Clear、Quit 或 Help semantic command，並確保顯示 label 與實際 key binding 來自相同定義。

#### Scenario: 執行附件指令

- **WHEN** 使用者依序按下 `Ctrl+X`、`a`
- **THEN** 系統執行 Attach command 並開啟現有 Filepicker Overlay

#### Scenario: 進入模板指令

- **WHEN** 使用者依序按下 `Ctrl+X`、`t`
- **THEN** 系統進入 template prefix，等待使用者選擇 HTML、Plain Text 或 EML template

#### Scenario: 顯示 Help

- **WHEN** 使用者依序按下 `Ctrl+X`、`?`
- **THEN** 系統顯示包含 direct shortcuts 與 prefix commands 的 Help overlay

### Requirement: 破壞性指令確認

系統 SHALL 在 Compose 含有任一 Header、Body 或附件內容時，要求使用者再次輸入相同 command key 才能執行 Clear 或 Quit；Esc SHALL 取消確認且保留所有內容。

#### Scenario: 確認清除非空 Compose

- **WHEN** Compose 含有內容，且使用者依序按下 `Ctrl+X`、`c`
- **THEN** 系統顯示 Clear 確認提示，且只有再次按下 `c` 才清除 Header、Body、Preview 與附件

#### Scenario: 取消清除

- **WHEN** 系統正等待 Clear 確認且使用者按下 Esc
- **THEN** 系統返回 normal 狀態並保留所有 Compose 內容

#### Scenario: 確認退出非空 Compose

- **WHEN** Compose 含有內容，且使用者依序按下 `Ctrl+X`、`q`
- **THEN** 系統顯示 Quit 確認提示，且只有再次按下 `q` 才終止 TUI

#### Scenario: 退出空白 Compose

- **WHEN** Compose 不含 Header、Body 或附件內容，且使用者依序按下 `Ctrl+X`、`q`
- **THEN** 系統不要求第二次確認並終止 TUI

### Requirement: Overlay 與 prefix 的按鍵隔離

系統 SHALL 優先將按鍵交給目前開啟的 overlay 或確認狀態，避免 prefix command 或背景 Compose 操作穿透。

#### Scenario: Filepicker 中按 Esc

- **WHEN** Filepicker Overlay 已開啟且使用者按下 Esc
- **THEN** 系統只關閉 Filepicker，返回原 Compose 焦點，不清除內容也不退出

#### Scenario: Prefix key 不寫入輸入欄位

- **WHEN** 使用者在任一 prefix 或確認狀態輸入 command key
- **THEN** 該按鍵只由 command dispatcher 處理，不得傳遞給 Header textinput 或 Composer textarea
