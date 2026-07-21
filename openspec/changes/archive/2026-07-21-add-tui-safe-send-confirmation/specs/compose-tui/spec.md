## ADDED Requirements

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

### Requirement: 安全確認期間快捷鍵隔離
ComposeModel MUST 在 pending confirmation state 優先處理確認按鍵，且 SHALL NOT 讓一般全域或 panel 快捷鍵在背景生效。

#### Scenario: 確認期間再次按 Ctrl+S
- **WHEN** 使用者在確認畫面再次按下 `Ctrl+S`
- **THEN** 系統不建立寄信 command，也不疊加第二個確認狀態

#### Scenario: 確認期間按附件或模板快捷鍵
- **WHEN** 使用者在確認畫面按下 `Ctrl+A`、`Ctrl+H`、`Ctrl+T` 或 `Ctrl+E`
- **THEN** 系統不開啟 Filepicker、不替換 Composer 內容，並維持相同 pending confirmation
