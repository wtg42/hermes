# Legacy Send Adapter Specification

## Purpose

定義 legacy Viper 郵件設定轉換至共用 `SMTPMailer` pipeline 時的資料契約、錯誤處理與相容行為。

## Requirements

### Requirement: Legacy 設定安全轉換為 MailCompose
系統 SHALL 將指定 Viper map 的 host、port、from、to、cc、bcc、subject、contents 與 attachment 轉換為 concrete `MailCompose`。必要欄位缺少或型別錯誤 MUST 回傳可辨識欄位的 error，MUST NOT panic 或開始 SMTP。

#### Scenario: 完整 legacy 設定
- **WHEN** map 包含所有有效字串欄位
- **THEN** 系統建立值與輸入一致的 MailCompose，並將 To/CC/BCC 表示為既有多地址資料

#### Scenario: Port 缺少或空白
- **WHEN** port 未提供或為空字串
- **THEN** MailCompose 使用 port 25

#### Scenario: 可選欄位缺少
- **WHEN** cc、bcc 或 attachment 未提供
- **THEN** 系統以空值建立 MailCompose，不回傳錯誤

#### Scenario: 必要欄位缺少或型別錯誤
- **WHEN** host、from、to、subject 或 contents 缺少或不是字串
- **THEN** 系統回傳指出欄位的 error，且不呼叫 Mailer／SMTP

### Requirement: Legacy send 沿用共用 SMTPMailer pipeline
`SendMailWithMultipart(key)` SHALL 在轉換成功後以 `SMTPMailer.Send` 執行地址驗證、附件載入、Header/MIME 組裝、envelope 建立與 SMTP 寄送。Legacy 函式 MUST NOT 維護另一套 MIME 或直接 SMTP 發送流程。

#### Scenario: Legacy send 成功
- **WHEN** legacy 設定與附件有效且 SMTP 接受郵件
- **THEN** 函式經共用 SMTPMailer 寄送一次並回傳 `(true, nil)`

#### Scenario: 共用 pipeline 失敗
- **WHEN** 地址、附件、MIME 或 SMTP 任一步驟回傳錯誤
- **THEN** 函式回傳 `(false, error)`，不執行第二次寄送

### Requirement: Legacy observable 郵件行為保持相容
重構後 legacy 入口 MUST 維持既有 port 預設、UTF-8 中文主旨、plain/HTML multipart、附件、可見 To/CC header 與只存在 envelope 的 BCC 行為。

#### Scenario: CC 與 BCC 郵件
- **WHEN** legacy 設定同時包含 To、CC 與 BCC
- **THEN** SMTP envelope 包含全部收件者，header 包含 To/CC 且不包含 BCC

#### Scenario: 中文與附件郵件
- **WHEN** Subject/Contents 包含中文且 attachment 有效
- **THEN** Mailpit 解析結果與重構前一致，包含中文內容與附件 MIME part

### Requirement: Legacy adapter 重構範圍受限
系統 MUST 保留 `SendMailWithMultipart` 名稱、Viper key input 與 `(bool, error)` contract，且 SHALL NOT 因本次重構改變 Burst、TUI、structured send、SMTP TLS/Auth policy 或 history。

#### Scenario: 既有呼叫方式
- **WHEN** 現有呼叫者以 Viper key 呼叫 `SendMailWithMultipart`
- **THEN** 呼叫方式無需修改並取得既有形式的成功或錯誤結果
