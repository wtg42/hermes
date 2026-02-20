## Context

目前 `sendmail` 模組包含多個郵件發送函數：
- `DirectSendMail()`: 用於 CLI 模式，只支援純文字
- `DirectSendMailFromTui()`: 廢棄的 TUI 函數
- `SendMailWithMultipart()`: 現有的 TUI 函數，支援 multipart 格式
- `BurstModeSendMail()`: Burst 模式，自行實現 MIME 構建

代碼重複問題：
- Header 設置、base64 編碼、email 驗證邏輯重複
- MIME multipart 構建邏輯在 `use_direct_send.go` 和 `burst_send.go` 中重複
- 錯誤處理不一致（有 panic、log 和 error 三種風格）
- 調試日誌遺留在代碼中
- 硬編碼值散落各處

## Goals / Non-Goals

**Goals:**
1. 統一所有郵件發送場景（CLI、TUI、Burst）使用同一函數
2. 消除重複代碼，提高可維護性
3. 統一錯誤處理方式
4. 清理技術債（調試日誌、廢棄代碼）
5. 提高可配置性（如 port 預設值）
6. 為補充測試奠定基礎

**Non-Goals:**
- 改變外部 API 行為（用戶視角無變化）
- 增加新的郵件發送功能
- 改造 CLI 參數系統（可規劃為後續 change）

## Decisions

### 1. 統一使用 SendMailWithMultipart()
**決定**: 所有郵件發送統一使用 `SendMailWithMultipart()` 函數。

**理由**:
- `SendMailWithMultipart()` 已支援純文字、HTML、附件等多種格式
- 是最完整的實現，可適應所有場景
- 避免維護多套邏輯

**替代方案考慮**:
- 創建新的抽象層（如 `MailBuilder` interface）：過度設計，現階段不需要
- 保留 `DirectSendMail()` 作簡單路徑：增加維護負擔

### 2. 移除 DirectSendMail() 和 CLI 直接發送
**決定**: 移除 `DirectSendMail()` 函數和 `cmd/direct_send_cmd.go`，CLI 使用改進為改造或規劃為未來工作。

**理由**:
- 用戶反饋 CLI 模式使用很少
- TUI 提供更好的用戶體驗
- 未來可規劃更友好的 CLI 參數系統

**替代方案考慮**:
- 改造 `direct_send_cmd.go` 使用 TUI 邏輯：可以，但超出本次範圍

### 3. 提取 MIME 構建邏輯為共用函數
**決定**: 將 MIME multipart 構建邏輯提取為 `buildEmailContent()` 等共用函數，供 `SendMailWithMultipart()` 和 Burst mode 使用。

**理由**:
- 當前 `SendMailWithMultipart()` 和 `BurstModeSendMail()` 都實現了相似的 MIME 構建
- 提取後易於測試、修改、維護

**具體做法**:
- 創建 `buildEmailHeaders()`: 構建郵件 header 部分
- 創建 `buildMIMEContent()`: 構建 multipart content 部分
- 兩者共用 `encodeRFC2047()` 和 email 驗證函數

### 4. 統一錯誤處理
**決定**: 所有郵件發送函數返回 `(bool, error)` 或 `error`，避免 panic 和忽略 error。

**理由**:
- 調用者可以優雅地處理錯誤
- 易於測試
- 一致的錯誤報告方式

**改變**:
- 移除 `panic()` 調用（如 `burst_send.go:84` 的 `panic(err)`）
- 改為返回 error
- 日誌記錄統一使用 `log.Printf()` 或結構化日誌（如果有的話）

### 5. 配置參數管理
**決定**: 保持使用 viper 進行配置，但明確默認值：
- SMTP 埠口預設值：25
- 字符集：UTF-8
- 編碼方式：base64

**理由**:
- Viper 已在項目中使用
- TUI 可通過 viper 動態配置參數
- 符合現有架構

## Risks / Trade-offs

### Risk 1: 破壞現有用戶的 CLI 腳本
**[Risk]** 移除 `direct_send_cmd.go` 可能破壞依賴 CLI 的自動化腳本
**[Mitigation]**
- 用戶反饋 CLI 使用很少，已確認可移除
- 文檔更新說明改用 TUI
- 未來可規劃改進的 CLI 介面

### Risk 2: Burst mode 性能
**[Risk]** 統一使用 `SendMailWithMultipart()` 後，Burst mode 性能可能下降
**[Mitigation]**
- SendMailWithMultipart() 已相當高效
- Burst mode 已使用 goroutine 和併發
- 如需優化，可後續進行性能評測和調整

### Risk 3: 測試覆蓋不足
**[Risk]** 重構後若無充分測試，可能引入新的 bug
**[Mitigation]**
- 本 change 明確包含補充測試任務
- 測試應覆蓋各種場景（純文字、HTML、附件、中文、多收件人）

### Trade-off: 代碼簡化 vs. 功能完整性
使用 multipart 格式發送純文字郵件稍微增加了複雜度（相對於簡單的 base64），但換取代碼統一和功能完整性，是可接受的權衡。

## Migration Plan

### 步驟 1: 代碼重構（保持向後相容）
1. 提取共用函數（`buildEmailHeaders()` 等）
2. 修改 `SendMailWithMultipart()` 使用共用函數
3. 修改 `BurstModeSendMail()` 使用共用函數
4. 移除廢棄代碼和調試日誌
5. 統一錯誤處理

### 步驟 2: CLI 模式調整
- 根據後續計畫決定是否移除或改造 `direct_send_cmd.go`
- 本次可標記為「廢棄中」或「規劃改進」

### 步驟 3: TUI 增強
- 在郵件類型選擇中新增「純文字」、「HTML」、「帶附件」等選項
- 確保各種模式都通過 `SendMailWithMultipart()` 工作

### 步驟 4: 測試驗證
- 編寫單元測試覆蓋各種郵件發送場景
- 手動測試 CLI、TUI、Burst mode
- 驗證中文編碼正常

### 步驟 5: 發布
- 更新文檔和 CHANGELOG
- 向用戶說明 CLI 變更（如適用）

## Open Questions

1. CLI 模式的未來計畫是什麼？是完全移除、改造還是計劃為另一個 change？
   - **當前決定**：由用戶決定，本次可標記為廢棄或規劃改進

2. 是否需要保留簡單的純文字發送路徑以提高性能？
   - **當前決定**：不需要，SendMailWithMultipart() 已足夠高效

3. Burst mode 是否有特殊的性能或功能要求？
   - **當前決定**：使用共用邏輯，如有性能問題後續改進
