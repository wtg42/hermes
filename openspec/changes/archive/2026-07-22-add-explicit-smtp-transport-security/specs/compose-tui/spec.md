## MODIFIED Requirements

### Requirement: 底部狀態列

系統 SHALL 在底部顯示狀態列，包含當下狀態可用的快捷鍵提示、prefix command 入口與使用者設定的 SMTP target。除非目前 model 已收到該連線完成 TLS 驗證的結果，狀態列 MUST NOT 宣稱 connected、TLS active 或 Auth 成功。

#### Scenario: Normal 狀態顯示精簡快捷鍵

- **WHEN** 撰寫畫面處於 normal 狀態且未在寄送
- **THEN** 底部狀態列顯示 `[Ctrl+S] Send`、`[Ctrl+X] Commands` 以及目前 panel 的導航提示

#### Scenario: Prefix 狀態顯示 Command HUD

- **WHEN** 使用者按下 `Ctrl+X` 進入 prefix command 模式
- **THEN** 底部狀態列暫時改為顯示 root Command HUD，而不是 normal 狀態快捷鍵

#### Scenario: 狀態列顯示 SMTP target

- **WHEN** 使用者填入 Host 和 Port，但目前 model 沒有已驗證的連線結果
- **THEN** 底部狀態列顯示例如 `SMTP target smtp.example.com:587`，不顯示 `Connected` 或 `TLS active`

#### Scenario: 狀態列實時更新 target

- **WHEN** 使用者修改 Host 或 Port 欄位
- **THEN** 狀態列的 SMTP target 即時更新，且不把欄位值視為連線成功證據
