## 1. Mailpit 容器生命週期管理

- [x] 1.1 修改 Makefile 的 test 目標，在啟動容器後添加 API 可用性檢查
- [x] 1.2 確保容器在所有測試完成後才清理
- [x] 1.3 測試 `make test` 命令能正確啟動和清理 Mailpit 容器

## 2. Mailpit API 客戶端實現

- [x] 2.1 在 sendmail 包中創建 `mailpit_client.go`，實現 Mailpit REST API 客戶端
- [x] 2.2 實現 `getLatestMessage()` 函數，從 API 獲取最新郵件
- [x] 2.3 實現 `searchMessages(query)` 函數，搜尋特定主題的郵件
- [x] 2.4 實現 `getRawMessage(id)` 函數，獲取郵件原始內容
- [x] 2.5 添加 HTTP 超時設置（5 秒）和錯誤處理

## 3. 郵件內容驗證輔助函數

- [x] 3.1 創建 `email_assertion.go`，提供郵件內容驗證的輔助函數
- [x] 3.2 實現 `assertSubjectEquals()` 驗證郵件主題
- [x] 3.3 實現 `assertFromEquals()` 驗證發件人
- [x] 3.4 實現 `assertToContains()` 驗證收件人列表
- [x] 3.5 實現 `assertContentContains()` 驗證郵件正文內容
- [x] 3.6 實現 `assertAttachmentExists()` 驗證附件存在

## 4. 集成測試 - 郵件內容驗證

- [x] 4.1 在 `integration_test.go` 中添加 `TestIntegrationEmailContentVerification` 測試
  - 驗證郵件主題是否被正確發送
  - 驗證郵件正文內容
  - 驗證 From、To 字段
- [x] 4.2 添加 `TestIntegrationComplexEmailContent` 測試
  - 驗證多個 To、Cc 收件人
  - 確保 BCC 不出現在郵件頭

## 5. 集成測試 - 字符編碼驗證

- [x] 5.1 在 `integration_test.go` 中添加 `TestIntegrationChineseEncoding` 測試
  - 使用中文主題發送郵件
  - 驗證主題被正確編碼和解碼
- [x] 5.2 添加 `TestIntegrationChineseContent` 測試
  - 驗證郵件正文中的中文內容被正確發送
  - 確保 base64 編碼/解碼正確

## 6. 集成測試 - 附件驗證

- [x] 6.1 在 `integration_test.go` 中添加 `TestIntegrationAttachmentInEmail` 測試
  - 創建測試附件文件
  - 發送包含附件的郵件
  - 驗證附件在郵件中存在
  - 驗證附件文件名正確

## 7. 集成測試 - MIME 結構驗證

- [x] 7.1 在 `integration_test.go` 中添加 `TestIntegrationMIMEStructure` 測試
  - 驗證郵件包含正確的 MIME 邊界
  - 驗證各部分的 Content-Type 正確

## 8. 集成測試 - 爆發模式

- [x] 8.1 在 `integration_test.go` 中添加 `TestIntegrationBurstModeSample` 測試
  - 使用爆發模式發送少量郵件（10 封）
  - 驗證所有郵件都被 Mailpit 接收
  - 檢查郵件計數是否正確

## 9. 驗證和測試

- [x] 9.1 執行 `make test` 確保所有集成測試通過
- [x] 9.2 驗證郵件驗證邏輯的準確性
- [x] 9.3 檢查測試覆蓋率，確保 sendmail 包的覆蓋達到目標
- [x] 9.4 測試 Docker 不可用的情況，確保有清晰的錯誤信息

## 10. 文檔和清理

- [x] 10.1 在 README 或開發文檔中記錄集成測試的使用方式
- [x] 10.2 說明 Mailpit API 驗證的工作原理
- [x] 10.3 清理臨時測試文件（如測試附件）
