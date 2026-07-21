## ADDED Requirements

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
