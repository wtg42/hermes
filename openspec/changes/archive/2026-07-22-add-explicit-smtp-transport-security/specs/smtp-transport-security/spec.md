## ADDED Requirements

### Requirement: SMTP transport 使用明確資料設定
系統 SHALL 以具體資料設定描述 TCP connect server、port、TLS mode、TLS server name、Auth mode、username 與當次 password，並 MUST 將 transport 設定與郵件 message／envelope 資料分離。TLS mode MUST 僅接受 `none` 或 `required`；Auth mode MUST 僅接受 `none` 或 `plain`。

#### Scenario: 未指定安全傳輸選項
- **WHEN** structured send 未指定 TLS 或 Auth 參數
- **THEN** 系統建立 `none` TLS 與 `none` Auth 的 transport 設定，維持既有明文 SMTP 行為

#### Scenario: 不支援的 transport mode
- **WHEN** 使用者指定不在封閉集合內的 TLS 或 Auth mode
- **THEN** 系統在讀取密碼或建立 TCP 連線前拒絕寄送

### Requirement: Required TLS 不得降級
當 TLS mode 為 `required` 時，系統 MUST 在 SMTP server 宣告 STARTTLS 後完成 TLS 升級與標準憑證驗證，並 MUST 使用明確 TLS server name 驗證 server identity。未宣告 STARTTLS、握手失敗或憑證驗證失敗時 SHALL 關閉連線且不得降級為明文寄送。

#### Scenario: STARTTLS 與憑證驗證成功
- **WHEN** server 宣告 STARTTLS、TLS handshake 成功且 certificate 符合指定 TLS server name
- **THEN** 系統在加密連線上繼續後續 SMTP 流程

#### Scenario: Server 未宣告 STARTTLS
- **WHEN** TLS mode 為 `required` 但 server 未宣告 STARTTLS
- **THEN** 系統回傳 TLS 階段錯誤，且不送出 MAIL、RCPT 或 DATA

#### Scenario: 憑證名稱或信任驗證失敗
- **WHEN** TLS handshake 的 certificate 不受信任、過期或不符合 TLS server name
- **THEN** 系統關閉連線並回傳 TLS 驗證錯誤，不降級且不傳送郵件

### Requirement: PLAIN Auth 僅能在已驗證 TLS 上執行
Auth mode 為 `plain` 時，系統 MUST 要求 TLS mode `required`、非空 username 與當次非空 password，並 SHALL 只在 TLS 成功後執行 SMTP PLAIN Auth。Auth 失敗時 MUST 在 MAIL／RCPT／DATA 前停止。

#### Scenario: TLS 上 PLAIN Auth 成功
- **WHEN** required TLS 已完成且 server 接受本次 username 與 password
- **THEN** 系統完成 Auth 後才傳送 envelope 與 message

#### Scenario: 明文 SMTP 搭配 PLAIN Auth
- **WHEN** Auth mode 為 `plain` 但 TLS mode 為 `none`
- **THEN** 系統在讀取密碼或連線前拒絕不安全的參數組合

#### Scenario: PLAIN Auth 被拒絕
- **WHEN** TLS 已完成但 server 拒絕 username 或 password
- **THEN** 系統回傳 Auth 階段錯誤並關閉連線，且不送出 MAIL、RCPT 或 DATA

### Requirement: Transport 失敗採 fail-closed 且錯誤可辨識
系統 SHALL 依 TCP、SMTP greeting／EHLO、STARTTLS、TLS verification、Auth 與 envelope／DATA 階段執行寄送。任一階段失敗 MUST 停止後續階段；回傳錯誤 SHALL 可辨識失敗階段，但 MUST NOT 包含 password 或完整 credential。

#### Scenario: TLS 前連線失敗
- **WHEN** TCP dial 或 SMTP greeting 失敗
- **THEN** 系統回傳對應階段錯誤且不嘗試 Auth 或 envelope

#### Scenario: 錯誤輸出不洩漏密碼
- **WHEN** 任一 transport 階段以包含底層 server response 的錯誤結束
- **THEN** 使用者可辨識失敗階段，但輸出不包含本次 password

#### Scenario: Transport 成功完成
- **WHEN** 設定要求的 TLS 與 Auth 均成功且 SMTP server 接受 envelope 與 DATA
- **THEN** 系統只送出一次郵件並回傳成功
