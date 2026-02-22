## ADDED Requirements

### Requirement: 幫助訊息不執行圖像繪製

系統在顯示幫助訊息時（用戶執行 `-h` 或 `--help` 旗標），應跳過 gopher ASCII art 圖像的繪製和輸出，以提升幫助功能的響應速度和使用者體驗。

#### Scenario: 使用 -h 旗標顯示幫助

- **WHEN** 使用者執行 `hermes -h`
- **THEN** 系統顯示幫助訊息，並跳過 gopher 圖像繪製和輸出

#### Scenario: 使用 --help 旗標顯示幫助

- **WHEN** 使用者執行 `hermes --help`
- **THEN** 系統顯示幫助訊息，並跳過 gopher 圖像繪製和輸出

#### Scenario: 執行有效命令後繪製圖像

- **WHEN** 使用者執行有效命令或預設命令（如 `hermes` 或 `hermes start-tui`）
- **THEN** 系統執行命令，命令完成後繪製並輸出 gopher ASCII art 圖像

#### Scenario: 子命令幫助（如 burst -h）

- **WHEN** 使用者執行子命令幫助（如 `hermes burst -h`）
- **THEN** 系統顯示子命令幫助訊息，並跳過 gopher 圖像繪製

### Requirement: Execute 函式返回幫助狀態

`Execute()` 函式應返回一個布爾值，指示程式是否顯示了幫助訊息。

#### Scenario: 顯示幫助時返回 false

- **WHEN** 使用者執行含有幫助旗標的命令
- **THEN** `Execute()` 返回 `false`，表示應跳過圖像繪製

#### Scenario: 正常執行時返回 true

- **WHEN** 使用者執行不含幫助旗標的有效命令
- **THEN** `Execute()` 返回 `true`，表示應繼續執行圖像繪製

#### Scenario: 發生錯誤時返回 true

- **WHEN** 使用者執行無效命令或發生錯誤
- **THEN** `Execute()` 返回 `true`，圖像繪製邏輯應執行（或由 main 中的 error handling 處理）
