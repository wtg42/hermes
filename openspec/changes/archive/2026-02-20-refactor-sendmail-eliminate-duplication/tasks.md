## 1. 分析與準備

- [x] 1.1 審查現有的 sendmail 模組代碼，確認重複邏輯
- [x] 1.2 檢查所有調用 DirectSendMail() 和 DirectSendMailFromTui() 的地方
- [x] 1.3 確認 direct_send_cmd.go 的使用情況

## 2. 提取共用函數

- [x] 2.1 創建 buildEmailHeaders() 函數，統一處理 From, To, Cc, Subject, MIME-Version 等 header
- [x] 2.2 創建 buildMIMEContent() 函數，統一處理 multipart content 構建
- [x] 2.3 整合現有的 encodeRFC2047() 函數，確保可被所有模塊使用
- [x] 2.4 整合 ValidateEmails() 函數，確保 email 驗證邏輯統一

## 3. 重構 sendmail 模組

- [x] 3.1 修改 SendMailWithMultipart()，改用 buildEmailHeaders() 和 buildMIMEContent()
- [x] 3.2 修改 BurstModeSendMail()，改用共用函數而非自行實現 MIME 構建
- [x] 3.3 統一錯誤處理：移除所有 panic()，改為返回 error
- [x] 3.4 移除廢棄的 DirectSendMailFromTui() 函數
- [x] 3.5 移除或重構 DirectSendMail() 函數（可根據後續計畫標記為廢棄或直接移除）
- [x] 3.6 刪除調試日誌（如 "tttttt=>" 等）

## 4. 清理和配置

- [x] 4.1 確認 port 預設值為 25，通過配置而非硬編碼
- [x] 4.2 檢查並移除其他硬編碼的魔法值
- [x] 4.3 更新 use_direct_send.go 和 burst_send.go 的註釋，反映新的架構
- [x] 4.4 確保所有函數有清晰的文檔註釋

## 5. CLI 和 TUI 集成

- [x] 5.1 檢查 cmd/direct_send_cmd.go，決定是否移除或改造
- [x] 5.2 檢查 TUI 中 SendMailWithMultipart() 的調用，確保工作正常
- [ ] 5.3 在 TUI 中新增郵件類型選擇（純文字、HTML、帶附件）
- [x] 5.4 測試 Burst mode 的郵件發送流程

## 6. 測試補齊

- [x] 6.1 編寫 buildEmailHeaders() 的單元測試（通過現有測試覆蓋）
- [x] 6.2 編寫 buildMIMEContent() 的單元測試（通過現有測試覆蓋）
- [x] 6.3 編寫 SendMailWithMultipart() 的集成測試（純文字、HTML、附件場景）
- [x] 6.4 編寫 BurstModeSendMail() 的测试
- [x] 6.5 測試中文主題和內容的編碼（RFC 2047 和 base64）
- [x] 6.6 測試多收件人（to, cc, bcc）的正確處理
- [x] 6.7 測試無效 email 地址的過濾

## 7. 驗證與文檔

- [x] 7.1 運行 `make test` 確保所有測試通過
- [x] 7.2 運行 `make lint` 檢查代碼風格
- [ ] 7.3 手動測試 TUI 的各種郵件發送模式
- [ ] 7.4 手動測試 Burst mode
- [ ] 7.5 更新 README.md 或相關文檔
- [ ] 7.6 更新 CHANGELOG（如存在）

## 8. 回顧與整理

- [x] 8.1 代碼審查：確認沒有遺漏重複代碼
- [x] 8.2 確認錯誤處理一致且正確
- [x] 8.3 檢查是否還有廢棄代碼或調試日誌
- [x] 8.4 確認所有改動符合設計文檔
