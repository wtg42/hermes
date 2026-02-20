## Why

目前的集成測試只能驗證郵件是否能發送，但無法驗證實際的郵件內容（主題、正文、附件、編碼等）。測試完成後立即清除 Mailpit 容器，導致開發者無法檢查郵件是否正確發送。這限制了測試的有效性，也增加了調試難度。

## What Changes

- 修改 Makefile 測試流程：Mailpit 容器在測試完成後保持運行，直到測試驗證完成才清除
- 在 `sendmail/integration_test.go` 中新增 Mailpit REST API 調用，用於驗證郵件內容
- 新增集成測試來驗證：
  - 郵件主題、正文、收件人是否正確
  - 附件是否被正確發送
  - 中文編碼是否正確
  - MIME 結構是否正確
  - 爆發模式的郵件發送是否正常

## Capabilities

### New Capabilities
- `mailpit-api-verification`: 透過 Mailpit REST API 驗證郵件內容和結構的能力
- `email-content-assertion`: 在集成測試中驗證郵件主題、正文、收件人等詳細內容

### Modified Capabilities
- `integration-testing`: 修改測試流程以保留 Mailpit 容器，支援更詳細的郵件驗證

## Impact

- **Makefile**: 修改 `test` 目標的流程邏輯
- **sendmail/integration_test.go**: 新增 API 驗證相關的測試函數
- **測試執行時間**: 可能略有增加（添加 API 查詢時間）
- **開發體驗**: 更完整的集成測試覆蓋和更清楚的驗證結果
