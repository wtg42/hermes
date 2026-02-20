## Context

目前 `SendMailWithMultipart()` 函數在驗證 To/Cc/Bcc 時：
- **To 字段** — 驗證失敗時返回 error（正確）
- **Cc/Bcc 字段** — 直接丟棄無效地址，沒有任何警告或錯誤反饋（問題所在）

使用者無法立即知道自己輸入的地址是否有效，可能導致郵件發送到錯誤的收件人。

## Goals / Non-Goals

**Goals:**
- 統一 Cc/Bcc 驗證邏輯，無效地址時返回錯誤而不是默默過濾
- 提供清晰的錯誤訊息指出哪個字段有問題及具體原因
- 允許使用者看到錯誤後立即修正並重新嘗試發送

**Non-Goals:**
- 修改 `ValidateEmails()` 函數簽名（保持回傳 `([]string, []string)` 的相容性）
- 實現客戶端地址驗證（即客戶端是否真的存在）

## Decisions

**1. 驗證策略 — Cc/Bcc 有無效地址時停止發送**

決策：當 Cc 或 Bcc 包含任何無效地址時，函數返回 error，不發送郵件。

理由：
- 使用者輸入錯誤應立即反饋，而不是默默改變行為
- 避免發送給部分預期的收件人，造成使用者困惑
- 與 To 字段的處理保持一致性

替代方案考慮：
- 過濾無效地址並以警告形式顯示（被棄用，因為現有方案太被動）
- 允許使用者選擇是否繼續發送（增加 UI 複雜度，不必要）

**2. 錯誤訊息格式**

決策：錯誤訊息應列出具體的字段、無效地址及原因。

格式：
```
invalid addresses in 'cc': ["invalid@", "too-long@example.com"] (reason: not matching email pattern)
invalid addresses in 'bcc': ["bad@address"] (reason: not matching email pattern)
```

理由：
- 讓使用者快速定位問題
- 提供足夠信息以修正錯誤

**3. 函數回傳值**

決策：保持現有簽名 `(bool, error)`，錯誤時回傳 `false` 和具體 error 訊息。

理由：
- 相容現有呼叫方的代碼
- 清晰區分成功和失敗

## Risks / Trade-offs

**[Risk]** 使用者以前依賴於無效地址被默默過濾的行為

→ **Mitigation**: 這本身就是個 bug，改進是必要的。過渡期可加日誌警告，但不應保留舊行為。

**[Risk]** 現有的 CLI 或 TUI 呼叫可能未正確處理新的錯誤狀態

→ **Mitigation**: 需要檢查所有 `SendMailWithMultipart()` 的呼叫方，確保它們正確顯示錯誤訊息。

## Implementation Approach

1. **修改 `SendMailWithMultipart()`**
   - 將 Cc/Bcc 的驗證結果捕獲（不再丟棄）
   - 如果任一字段有無效地址，收集所有無效地址並返回詳細 error

2. **單元測試**
   - 驗證 Cc/Bcc 有無效地址時函數返回 error
   - 驗證錯誤訊息包含具體的無效地址和字段名

3. **呼叫方更新**
   - 檢查 TUI 和 CLI 中的錯誤處理
   - 確保錯誤訊息被正確顯示給使用者
