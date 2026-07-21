## ADDED Requirements

### Requirement: Structured send 必須使用明確 IPv4 server
系統 MUST 要求 `hermes send --server` 使用明確 dotted-decimal IPv4，且 SHALL NOT 從 hostname、localhost、IPv6、環境變數或應用設定推導 server。

#### Scenario: 有效 IPv4 server
- **WHEN** 使用者提供例如 `192.0.2.10` 的有效 IPv4
- **THEN** 系統接受該 server 並繼續其他預先驗證

#### Scenario: 缺少 server
- **WHEN** 使用者未提供 `--server`
- **THEN** 系統在寄信前回傳需要明確 IPv4 的錯誤

#### Scenario: hostname 或 localhost
- **WHEN** 使用者提供 hostname 或 `localhost`
- **THEN** 系統拒絕寄信，即使同時提供 `--confirm-outside-whitelist`

#### Scenario: IPv6 server
- **WHEN** 使用者提供 IPv6 address
- **THEN** 系統拒絕寄信，即使同時提供 `--confirm-outside-whitelist`

### Requirement: Structured send 安全白名單
系統 SHALL 將 `weitingshih@rd01.softnext.com.tw`、`jllee@rd01.softnext.com.tw`、`adam@rd01.softnext.com.tw` 視為安全寄件者，將收件網域精確等於 `rd01.softnext.com.tw` 視為安全收件範圍，並將 port 25 視為安全 port。比較 SHALL 不分大小寫且 MUST 採精確相等。

#### Scenario: 全部位於安全邊界
- **WHEN** From 位於安全寄件者清單、所有 To/CC/BCC 都屬於精確安全網域且 port 為 25
- **THEN** 系統不要求額外確認

#### Scenario: 同網域但寄件者不在白名單
- **WHEN** From 屬於 `rd01.softnext.com.tw` 但不是三個安全寄件者之一
- **THEN** 系統將 From 視為白名單外寄件者

#### Scenario: 收件者使用安全網域的子網域
- **WHEN** 任一 To/CC/BCC 使用 `mail.rd01.softnext.com.tw`
- **THEN** 系統將該收件者視為安全網域外

### Requirement: 白名單外寄信需要明確確認
若 From 不在安全寄件者清單、任一 To/CC/BCC 不在安全收件網域，或 port 不為 25，系統 MUST 要求 `--confirm-outside-whitelist`。系統 SHALL 彙整並回報所有確認原因。

#### Scenario: 白名單外寄件者未確認
- **WHEN** 使用者提供白名單外 From 且未提供確認旗標
- **THEN** 系統拒絕寄信並指出寄件者原因

#### Scenario: 外部收件人未確認
- **WHEN** 任一 To/CC/BCC 位於安全網域外且未提供確認旗標
- **THEN** 系統拒絕寄信並列出外部收件人

#### Scenario: 非 25 port 未確認
- **WHEN** 使用者指定非 25 port 且未提供確認旗標
- **THEN** 系統拒絕寄信並指出 port 原因

#### Scenario: 多個白名單外原因
- **WHEN** From、收件人與 port 同時超出安全邊界且未提供確認旗標
- **THEN** 系統在單一錯誤中列出全部原因，不送出郵件

#### Scenario: 明確確認白名單外寄信
- **WHEN** 所有格式驗證通過且使用者提供 `--confirm-outside-whitelist`
- **THEN** 系統允許本次白名單外 From、收件網域與 port 繼續發送

### Requirement: 確認旗標不得略過基本驗證
`--confirm-outside-whitelist` SHALL 只授權白名單邊界，MUST NOT 略過 server、Email、port 或附件格式與可讀性驗證。

#### Scenario: 確認旗標搭配無效 Email
- **WHEN** 使用者提供確認旗標但任一地址格式無效
- **THEN** 系統仍拒絕寄信並指出無效地址

#### Scenario: 確認旗標搭配不存在附件
- **WHEN** 使用者提供確認旗標但任一附件不存在或無法讀取
- **THEN** 系統仍拒絕寄信並指出附件路徑

### Requirement: Structured send 預先驗證必須零副作用
系統 MUST 在呼叫 Mailer 或 SMTP 前完成預設值、所有地址、server、port、安全確認與附件驗證。任一驗證失敗時 SHALL 不得送出部分郵件。

#### Scenario: 任一預先驗證失敗
- **WHEN** structured send 的任一輸入不符合需求
- **THEN** Mailer 與底層 SMTP 的呼叫次數皆為零
