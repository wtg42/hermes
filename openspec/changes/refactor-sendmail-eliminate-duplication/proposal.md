## Why

目前郵件發送模組存在大量代碼重複，多個函數實現了相似的 MIME 構建、base64 編碼、email 驗證邏輯。這導致維護困難、bug 難以修復、功能無法一致。通過統一郵件發送邏輯，可以降低複雜度、提高可維護性、為未來擴展奠定基礎。

## What Changes

- **統一郵件發送邏輯**：所有郵件發送（CLI、TUI、Burst mode）統一使用 `SendMailWithMultipart()` 函數，支援純文字、HTML、附件
- **移除重複實現**：刪除 `DirectSendMail()` 及已廢棄的 `DirectSendMailFromTui()` 函數
- **清理技術債**：
  - 移除調試日誌（如 `log.Println("tttttt=>")`）
  - 統一錯誤處理（用 `error` 返回值代替 `panic()`）
  - 提取硬編碼值為可配置變數（如 port 預設值 25）
- **改進 Burst mode**：使用共用的郵件構建邏輯而非自行實現
- **增強 TUI**：允許用戶選擇郵件類型（純文字、HTML、帶附件）
- **補充測試**：確保重構後的郵件發送邏輯有足夠的測試覆蓋

## Capabilities

### New Capabilities

無新功能，但改進既有功能的實現品質。

### Modified Capabilities

- `unified-email-sending`: 郵件發送邏輯的統一實現，支援多種格式（純文字、HTML、附件）並提供一致的 API

## Impact

- **sendmail/** 模組：重構代碼結構，合併重複函數
- **cmd/** 模組：可能移除 `direct_send_cmd.go` 或改造其使用的 API
- **tui/** 模組：增強郵件類型選擇功能
- **測試**：補充相關測試用例
- **無破壞性改動**：外部 API 行為保持一致
