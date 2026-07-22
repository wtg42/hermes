## Context

`SendMailWithMultipart` 從 `viper.GetStringMap(key)` 取得鬆散 map 後，以多個 type assertions 取值，並自行重做 `SMTPMailer.Send` 已有的地址驗證、附件載入、Header/MIME 組裝、envelope 建立與 SMTP 呼叫。這使修正 MIME 或 SMTP 時必須同步兩處，缺少 `to`、`cc`、`bcc` 等欄位還可能在 assertion 時 panic。

## Goals / Non-Goals

**Goals:**

- 將 legacy map 明確轉成 concrete `mail.MailCompose`。
- 所有欄位錯誤在 SMTP 前以 error fail-closed。
- Legacy 與現代入口共用 `SMTPMailer.Send` 的地址、附件、MIME 與 SMTP 行為。
- 保留 `(bool, error)` 與 Viper key 相容性。

**Non-Goals:**

- 移除 Viper、`SendMailWithMultipart` 或 `NewAttachmentLegacy`。
- 修改 Burst、TUI、structured send、MIME 格式或 transport policy。
- 建立 adapter/config interface、service 或 factory。

## Decisions

### 1. 以純資料轉換函式隔離 Viper map

新增未匯出的 `legacyMailCompose(map[string]any) (mail.MailCompose, error)`。它只解析 host、port、from、to、cc、bcc、subject、contents、attachment，沒有 I/O、Viper 存取或 SMTP 副作用。必填字串缺少／型別錯誤回傳包含欄位名稱的 error；port 缺少或空字串仍預設 25；cc、bcc、attachment 可缺少並視為空值。

地址字串維持原始 comma-separated 語意，轉成 slices 後交由 `SMTPMailer.Send` 的既有 `ValidateEmails` 規則處理，避免 adapter 複製驗證政策。

### 2. `SendMailWithMultipart` 成為薄型邊界

函式只執行 `GetStringMap`、轉換、`NewSMTPMailer().Send(compose)`，成功回傳 `(true, nil)`，任何錯誤回傳 `(false, err)`。不直接呼叫 `SendMail`、附件 loader 或 MIME builder。

### 3. 保留 Burst 使用的 helpers

`EmailData`、`buildEmailHeaders` 與 `buildMIMEContent` 仍被 Burst 使用，本次不刪除或改簽名。只有 legacy 函式內的重複流程移除，避免把 concurrency/random address 納入同一 change。

### 4. Characterization 優先

先以 table tests 固定每個 legacy 欄位的 required/optional/default/type 行為，再以 SMTP injection 與 Mailpit 驗證 To/CC/BCC envelope/header、附件、中文與 MIME output 等價。新增缺欄位／錯型別不得 panic 的測試。

## Risks / Trade-offs

- [SMTPMailer error wording與 legacy 略有差異] → Characterization 鎖定具意義欄位與 fail-closed 語意，不依賴無價值的完整字串；外部成功郵件內容必須一致。
- [StringSlice 解析改變空白語意] → 只做與既有 comma input 相同的切分，不新增 trim/normalization；地址合法性仍交給既有 validator。
- [Legacy 測試依賴全域 Viper/SendMail] → 保持現有 reset/injection 模式並新增純 converter tests，降低全域狀態覆蓋比例。
- [順手清理 Burst 擴大風險] → 明確保留 Burst helpers，後續另開 shared message builder change。

## Migration Plan

先補 characterization tests，再加入 converter，最後將 legacy 函式切到 SMTPMailer。完成 unit/race/Mailpit tests；history schema 與部署資料不需 migration，回復單一重構即可 rollback。

## Open Questions

無。Burst 共用 builder 留待下一個 change。
