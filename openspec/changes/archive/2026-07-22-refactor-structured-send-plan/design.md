## Context

`StructuredSender.Execute` 目前先建立 `MailCompose` 與 `SMTPTransportConfig`，再檢查注入的 `mail.Mailer` 是否剛好為 `*SMTPMailer`。符合時呼叫 `SendWithTransport`，其他實作則只收到 `MailCompose`，因此 TLS/Auth policy 會從執行資料流消失。這讓 production 與測試替身走不同分支，也迫使 history 從 compose 與額外 variadic transport 重新拼裝 record。

專案採資料導向設計；重構應讓一份 concrete plan 成為寄送 attempt、結果與 history 的共同輸入，而不是增加 transport hierarchy。

## Goals / Non-Goals

**Goals:**

- 在任何寄送副作用前建立完整 `StructuredSendPlan`。
- plan 明確包含 resolved `MailCompose` 與 ready `SMTPTransportConfig`。
- structured execution 只有一條 plan-aware attempt 路徑，不依 mailer concrete type 分流。
- execution result 保留實際使用的 plan，history 直接由同一 plan 建立。
- 保持 CLI 與儲存格式的既有 observable behavior。

**Non-Goals:**

- 建立通用 transport／repository／service interface。
- 改寫 `mail.Mailer` 或 TUI／EML 的寄送模型。
- 修改 SMTP protocol、TLS/Auth policy、MIME、history v2 schema 或 replay flags。
- 清理 `SendMailWithMultipart`、Viper legacy path、Burst 或拆分 Compose TUI。

## Decisions

### 1. 使用 concrete `StructuredSendPlan`

Plan 只包含：

```go
type StructuredSendPlan struct {
    Compose   mail.MailCompose
    Transport SMTPTransportConfig
}
```

它是一次 attempt 所需的完整資料快照。Plan 不保存 confirmation、`NoHistory` 或 password source control；但 `Transport.Password` 可在當次記憶體中存在，且既有 secret redaction／不持久化規則不變。

未把 transport 欄位塞進 `MailCompose`，因為 message/envelope 與連線認證資料仍是不同責任。

### 2. 將 plan build 與 attempt 分開

`StructuredSender` 依序執行：resolve defaults、格式驗證、安全評估、transport ready validation、附件 preflight、補隨機內容、建立 plan。完成 plan 前不得呼叫 attempt。

Attempt 使用小型函式值：

```go
type StructuredSendAttempt func(StructuredSendPlan) error
```

正式 command 注入包裝 `SMTPMailer.SendWithTransport(plan.Compose, plan.Transport)` 的函式；測試注入記錄完整 plan 的函式。這是 I/O 邊界函式，不是為多型建立 interface。

未保留 `NewStructuredSender(mail.Mailer)` 的 structured execution 相容 wrapper，因為 wrapper 必須丟棄 transport，正是本次要消除的問題。Repo 內 call sites 一次更新；CLI 與使用者資料格式不變。

### 3. Execution result 直接攜帶 plan

`StructuredSendExecution` 改為保存 `Plan`、`Attempted` 與 `AttemptError`。只有 plan build 完成且即將呼叫 attempt 時才設定 attempted。如此 preflight failure 與 attempt failure 的界線保持明確，history 仍只記錄實際 attempt。

### 4. History record 只接受 plan

History 建構函式從 plan 複製 message 與非敏感 transport metadata，並以 plan 中的當次 password 做錯誤 redaction，但永遠不序列化 password。移除「compose 加 optional variadic transport」入口，避免呼叫者忘記傳 transport 後產生與實際寄送不同的 record。

### 5. Replay 仍先重建 options，再建立新 plan

History replay 不直接重用保存的 plan，也不直接呼叫 attempt。Record 仍先轉為目前的 `StructuredSendOptions`，重新取得 password、執行現行驗證並建立新 plan，維持既有 fail-closed 與 fresh authorization 行為。

### 6. 以 characterization tests 鎖定等價行為

重構前先補測試證明：

- attempt 收到的 compose 與 transport 等於 resolved values；
- 非 SMTP concrete test attempt 仍收到 TLS/Auth policy；
- 所有 preflight failure 都不呼叫 attempt；
- history message/transport metadata 等於 attempt 使用的 plan；
- password 不進入 JSON/error output；
- no-history、send failure、history failure 各只 attempt 一次。

## Risks / Trade-offs

- [變更 exported Go constructor／history helper 的 source compatibility] → 這些 API 目前僅由同 repo 使用；一次更新所有 call sites，CLI 與資料格式保持不變，並用 `rg` 確認沒有漏網呼叫。
- [Plan 內暫存 password] → 保持當次短生命週期，不提供 JSON tags 或持久化 plan 的 API；history 只挑選非敏感欄位並執行 redaction。
- [函式注入可能被濫用成抽象層] → 僅保留單一 attempt signature，不建立 registry、factory 或多層 wrapper。
- [重構不慎改變 validation 順序] → 先以 characterization tests 鎖定錯誤優先順序、零副作用與隨機內容行為。
- [同時清理 legacy path 擴大風險] → 明確排除 `SendMailWithMultipart` 與 Viper，另開後續 change。

## Migration Plan

1. 先新增 plan/attempt characterization tests，維持現況失敗以證明 type assertion 問題。
2. 新增 plan data type 與 builder，讓既有 validation logic 逐步搬入而不改順序。
3. 將 structured production/test call sites 改成 plan-aware attempt。
4. 將 execution 與 history record 改以 plan 為唯一資料來源。
5. 執行 unit、race、Mailpit integration 與 strict spec validation；若失敗可回復此單一重構提交，history v2 無需 migration。

## Open Questions

無。Legacy mail path 與 TUI file split 留待後續獨立 change。
