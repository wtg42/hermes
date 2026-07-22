# direct-send-safety Specification

## Purpose
定義 structured 單封寄信在連線 SMTP 前必須遵守的 server、白名單、確認與零副作用安全邊界。

## Requirements

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

### Requirement: 單封寄信共用安全評估

系統 SHALL 以同一個單封安全政策評估 structured CLI 與 Compose TUI 的最終 From、To、CC、BCC 與 port。安全政策 MUST 使用既有三個安全寄件者、精確 `rd01.softnext.com.tw` 收件網域及 port 25，並 SHALL 以穩定順序彙整所有越界原因。

#### Scenario: TUI 郵件全部位於安全邊界

- **WHEN** Compose TUI 的 From 位於安全寄件者清單、所有 To/CC/BCC 都屬於精確安全網域且 port 為 25
- **THEN** 安全評估不回傳越界原因，TUI 可直接進入既有寄信流程

#### Scenario: TUI 郵件包含多個越界原因

- **WHEN** Compose TUI 的 From、任一收件人與 port 同時超出安全邊界
- **THEN** 安全評估一次回傳全部原因，且每個外部 To/CC/BCC 都可由畫面識別

#### Scenario: Structured CLI 沿用共用政策

- **WHEN** `hermes send` 評估相同的 From、To、CC、BCC 與 port
- **THEN** 系統維持既有白名單、精確網域比較、原因彙整與確認旗標行為

#### Scenario: 單封安全評估收到無效地址或 port

- **WHEN** 任一 From、To、CC、BCC 格式無效或 port 不在有效範圍
- **THEN** 系統回傳基本格式錯誤而非確認原因，且不得以白名單確認繼續寄信

### Requirement: TUI 白名單外寄信需要文字確認

Compose TUI MUST 在白名單外寄信前顯示所有安全原因並要求使用者輸入大小寫完全相符的 `SEND`。進入確認狀態、輸入錯誤 token 或取消時 SHALL NOT 呼叫 Mailer。

#### Scenario: 白名單外郵件進入確認狀態

- **WHEN** 使用者在 Compose TUI 對含有越界原因的郵件按下 `Ctrl+S`
- **THEN** 系統顯示所有原因與 `SEND` 提示，保留撰寫內容且 Mailer 呼叫次數為零

#### Scenario: 輸入精確 SEND

- **WHEN** 使用者在確認狀態輸入 `SEND` 並按 Enter，且郵件快照仍一致
- **THEN** 系統只授權該封郵件進入既有非同步寄信流程一次

#### Scenario: 輸入錯誤確認文字

- **WHEN** 使用者輸入 `send`、其他文字或空值並按 Enter
- **THEN** 系統維持確認狀態、顯示錯誤提示且 Mailer 呼叫次數為零

#### Scenario: 取消白名單外寄信

- **WHEN** 使用者在確認狀態按 Esc
- **THEN** 系統取消 pending confirmation、返回原撰寫畫面、保留所有郵件內容且不呼叫 Mailer

#### Scenario: 確認不得重用

- **WHEN** 一次確認已成功、取消或寄信流程已結束
- **THEN** 系統清除該次授權，後續白名單外寄信必須重新輸入 `SEND`

### Requirement: TUI 確認綁定郵件快照

TUI 白名單外授權 MUST 只適用於產生警告時的 From、To、CC、BCC、Subject、Body、Host、Port 與附件快照。確認期間 SHALL 凍結底層郵件編輯與其他寄信快捷鍵，送出前 MUST 再次驗證快照一致性。

#### Scenario: 確認期間按下編輯快捷鍵

- **WHEN** pending confirmation 存在且使用者輸入一般編輯鍵、附件快捷鍵或再次按 `Ctrl+S`
- **THEN** 系統只更新或處理確認輸入，不修改底層郵件也不啟動第二次寄信

#### Scenario: 待確認快照與目前郵件不一致

- **WHEN** 系統在接受 `SEND` 前發現任一受保護欄位或附件已變更
- **THEN** 系統拒絕使用舊授權、不呼叫 Mailer，並要求對新內容重新進行安全評估

### Requirement: TUI 確認不得略過 Mailer fail-closed 驗證

文字 `SEND` SHALL 只解除白名單限制，MUST NOT 略過附件存在性、可讀性、MIME 組裝或其他 Mailer 驗證。

#### Scenario: 已確認郵件包含不存在附件

- **WHEN** 使用者正確確認白名單外郵件，但附件在 Mailer 驗證時不存在或無法處理
- **THEN** 系統顯示附件錯誤且底層 SMTP 呼叫次數為零

### Requirement: History replay 必須取得新的安全授權
History record MUST NOT 保存或重用先前的白名單確認。每次 replay SHALL 以 record 的 server、port、From、To、CC、BCC 與附件重新執行目前的 structured 單封驗證及安全政策。

#### Scenario: Replay 安全範圍內紀錄
- **WHEN** record 的 server、地址、port 與附件仍符合目前所有驗證及安全邊界
- **THEN** replay 不要求額外確認並可繼續呼叫 Mailer

#### Scenario: Replay 白名單外紀錄但未重新確認
- **WHEN** record 的 From、任一收件人或 port 位於目前白名單外，且本次 replay 未提供 `--confirm-outside-whitelist`
- **THEN** 系統列出目前全部越界原因並拒絕寄送，Mailer 與 SMTP 呼叫次數皆為零

#### Scenario: Replay 白名單外紀錄並重新確認
- **WHEN** record 所有基本驗證通過且本次 replay 明確提供 `--confirm-outside-whitelist`
- **THEN** 系統只授權本次 replay 的白名單外邊界並繼續寄送

#### Scenario: Replay 確認不得略過基本驗證
- **WHEN** 本次 replay 提供確認旗標但 record 的 server、Email、port 或附件不再有效
- **THEN** 系統仍在 SMTP 前拒絕寄送，且不建立新的 history record

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
