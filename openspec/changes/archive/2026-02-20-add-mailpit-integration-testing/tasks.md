## 1. 環境準備與配置

- [x] 1.1 創建 `docker-compose.yml` 在專案根目錄，定義 Mailpit 服務（SMTP 端口 1025）
- [x] 1.2 驗證 docker-compose.yml 語法正確（執行 `docker-compose config`）
- [x] 1.3 測試手動啟動容器（`docker-compose up -d` 和 `docker-compose down`）

## 2. 修改 Makefile 自動化測試流程

- [x] 2.1 在 Makefile 的 `test` target 前添加 Docker 可用性檢查
- [x] 2.2 修改 `test` target 在測試前啟動 Mailpit（`docker-compose up -d`）
- [x] 2.3 修改 `test` target 在測試後關閉 Mailpit（`docker-compose down`）
- [x] 2.4 添加錯誤處理：若 Docker 不可用，顯示提示訊息並中斷
- [x] 2.5 添加簡單延遲（sleep 2）確保 Mailpit 完全啟動後再執行測試

## 3. 環境變數與連接配置

- [x] 3.1 在測試代碼中添加環境變數支援（`TEST_SMTP_HOST`、`TEST_SMTP_PORT`）
- [x] 3.2 設置預設值為 `localhost:1025`
- [x] 3.3 驗證現有測試代碼能讀取這些變數（可在 `use_direct_send.go` 或相關文件中修改）

## 4. 集成測試實現

- [x] 4.1 創建新的集成測試文件（如 `sendmail/integration_test.go`）
- [x] 4.2 實現測試：發送簡單郵件到 Mailpit，驗證 smtp.SendMail 返回 nil
- [x] 4.3 實現測試：發送包含 To、Cc、Bcc 的複雜郵件到 Mailpit
- [x] 4.4 驗證現有 mock 測試仍然通過（不受 Mailpit 影響）

## 5. 驗收與清理

- [x] 5.1 執行 `make test` 確保完整流程成功運行
- [x] 5.2 驗證 `make test` 後沒有遺留容器（執行 `docker-compose ps`）
- [x] 5.3 測試 Docker 不可用場景（模擬無 Docker 環境）
- [x] 5.4 運行所有現有測試，確保無回歸
