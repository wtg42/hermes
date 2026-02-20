## Context

目前的集成測試架構：
- Mailpit 以 Docker 容器運行（通過 docker-compose.yml）
- 測試使用 SMTP 協議將郵件發送到 localhost:1025
- 測試完成後立即清除容器，導致無法驗證郵件內容

新增的 Mailpit 功能：
- 提供 REST API 在 http://localhost:8025/api/v1
- 支援查詢、搜尋、獲取郵件原始內容
- 支援訪問郵件的詳細信息（主題、收件人、附件等）

## Goals / Non-Goals

**Goals:**
1. 在集成測試中透過 API 驗證郵件的內容和結構
2. 驗證中文編碼、附件、MIME 結構等細節
3. 為爆發模式和直接發送功能添加集成測試覆蓋
4. 保持測試自動化，測試完成後自動清理環境

**Non-Goals:**
- 修改 SMTP 發送邏輯（use_direct_send.go、burst_send.go）
- 修改郵件構建邏輯（header、MIME 部分構建）
- 添加生產代碼依賴（API 驗證僅在測試中使用）
- 支援外部 Mailpit 實例（本地開發和 CI 測試）

## Decisions

### 決策 1：在測試中呼叫 Mailpit API 而非解析 SMTP 協議
**選擇：使用 REST API**

**理由：**
- Mailpit API 提供高級別的郵件信息訪問（不需低級 SMTP 操作）
- 簡化測試代碼，無需解析 MIME 結構
- 測試更易於理解和維護

**替代方案：**
- 直接解析 SMTP 回應：複雜，需要處理 MIME 細節
- 訪問容器內部文件系統：不適用於容器化環境

### 決策 2：Mailpit 容器在所有測試期間保持運行
**選擇：延遲清理直到所有驗證完成**

**理由：**
- 允許多個測試共享同一個 Mailpit 實例
- 減少容器啟動開銷
- 支援複雜的多郵件驗證場景

**替代方案：**
- 每個測試啟動清理一個容器：開銷大，難以進行跨測試驗證

### 決策 3：API 調用的超時和重試策略
**選擇：
- HTTP 連接超時：5 秒
- 不進行重試（郵件應立即出現在 Mailpit）
- 測試失敗時返回清晰的錯誤信息

**理由：**
- Mailpit 在本地運行，郵件延遲應該最小
- 簡化測試邏輯，避免不必要的複雜性

### 決策 4：在 Makefile 中分離驗證邏輯
**選擇：保持單一 `make test` 命令，但在幕後分離驗證步驟**

**流程：**
1. 啟動 Mailpit 容器
2. 執行 Go 測試（包括基本功能和 API 驗證）
3. 清理容器

**理由：**
- 對用戶透明，使用體驗不變
- 測試和驗證在同一過程中完成

## Risks / Trade-offs

### 風險 1：API 連接失敗
**風險：** 如果 Mailpit API 不可達，測試會失敗
**緩解策略：**
- 明確的錯誤信息指示 Mailpit 未運行
- 在測試開始時檢查 API 可用性
- 文檔中說明系統要求

### 風險 2：郵件信息不同步
**風險：** 郵件發送後 API 還未完全索引
**緩解策略：**
- Mailpit 郵件接收應該是即時的
- 如需要，可在測試中添加短暫延遲（1-2 秒）

### 風險 3：測試執行時間增加
**權衡：** API 查詢會增加測試時間（每次查詢 100-200ms）
**取捨：**
- 額外的驗證帶來更高的信心
- 集成測試本身已較慢（SMTP 操作）
- 增加的時間可接受（估計 +10-20% 整體時間）

### 風險 4：容器清理失敗導致殭屍進程
**風險：** docker-compose down 失敗時容器可能留下
**緩解策略：**
- 使用 `docker-compose down --remove-orphans`
- 在 Makefile 中檢查清理成功
- CI/CD 流程中設置容器重用政策

## Implementation Approach

### 修改 Makefile
```
test:
  1. 檢查 Docker 可用性
  2. 啟動 Mailpit (docker-compose up -d)
  3. 等待 API 可用 (檢查 http://localhost:8025/api/v1/messages)
  4. 執行測試 (go test -tags integration)
  5. 清理容器 (docker-compose down)
```

### 新增測試輔助函數
- `getLatestMessage()`: 從 API 獲取最新郵件
- `getMessageBySubject(subject)`: 搜尋特定主題的郵件
- `assertEmailContent(email, expectedSubject, expectedFrom, expectedTo)`: 驗證郵件內容

### 新增集成測試
- `TestIntegrationEmailContentVerification`: 驗證郵件主題和正文
- `TestIntegrationEmailAttachment`: 驗證附件
- `TestIntegrationEmailEncoding`: 驗證中文編碼
- `TestIntegrationBurstMode`: 驗證爆發模式發送多封郵件

## Open Questions

1. 是否需要支援 Mailpit 的 Basic Auth？（目前未配置）
2. 對於大量郵件（爆發模式），API 查詢性能如何？
3. 是否需要清理 Mailpit 的先前郵件數據（測試隔離）？
