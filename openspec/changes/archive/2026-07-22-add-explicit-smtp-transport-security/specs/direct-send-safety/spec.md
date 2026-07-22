## ADDED Requirements

### Requirement: Structured transport 參數必須在連線前完整驗證
系統 MUST 在讀取 password 或呼叫 Mailer／SMTP 前驗證 TLS mode、TLS server name、Auth mode、username 與 password source 的組合。`--confirm-outside-whitelist` SHALL NOT 略過任何 transport 驗證，且 TLS server name MUST NOT 取代或重新解析明確的 `--server` IPv4 連線目標。

#### Scenario: Required TLS 缺少 server name
- **WHEN** TLS mode 為 `required` 但未提供有效的 TLS server name
- **THEN** 系統在讀取 stdin 或連線前拒絕寄送

#### Scenario: TLS server name 不改變連線目標
- **WHEN** 使用者提供明確 IPv4 server 與不同的 TLS server name
- **THEN** 系統只連線至該 IPv4，並只將 TLS server name 用於 certificate identity verification

#### Scenario: 無 TLS 時提供 TLS server name
- **WHEN** TLS mode 為 `none` 但使用者提供 TLS server name
- **THEN** 系統拒絕無作用且可能造成誤解的參數組合

#### Scenario: 確認旗標搭配無效 Auth 組合
- **WHEN** 使用者提供 `--confirm-outside-whitelist`，但 PLAIN Auth 缺少 required TLS、username 或 password source
- **THEN** 系統仍在讀取 password 與連線前拒絕寄送
