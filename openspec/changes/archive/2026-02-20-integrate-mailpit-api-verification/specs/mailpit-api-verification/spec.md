# Mailpit API 驗證

## Purpose

透過 Mailpit 提供的 REST API 驗證發送出去的郵件內容和結構，確保郵件的正確性。

## Requirements

### Requirement: Mailpit API 連接和查詢功能
測試代碼應能連接到 Mailpit 的 REST API，並查詢接收到的郵件。

#### Scenario: 成功連接到 Mailpit API
- **WHEN** 測試代碼啟動且 Mailpit 容器運行中
- **THEN** 能成功連接到 `http://localhost:8025/api/v1/messages`

#### Scenario: 查詢最新郵件
- **WHEN** 調用 `GET /api/v1/message/latest`
- **THEN** 返回最新發送的郵件的詳細信息（包含 ID、Subject、From、To 等）

#### Scenario: 查詢郵件原始內容
- **WHEN** 調用 `GET /api/v1/message/{ID}/raw`
- **THEN** 返回完整的郵件原始格式（包含所有 MIME 部分和附件）

### Requirement: API 客戶端實現
提供能在集成測試中調用 Mailpit API 的客戶端函數。

#### Scenario: 查詢郵件的輔助函數存在
- **WHEN** 測試代碼需要查詢郵件
- **THEN** 存在 `getLatestMessage()` 等輔助函數簡化 API 調用

#### Scenario: 解析郵件信息
- **WHEN** 從 API 獲取郵件數據
- **THEN** 能解析並提取主題、發件人、收件人等信息供測試斷言使用

### Requirement: API 錯誤處理
當 Mailpit API 不可用或返回錯誤時，測試應能妥善處理。

#### Scenario: API 連接失敗
- **WHEN** Mailpit 容器未運行或 API 不可達
- **THEN** 測試返回明確的錯誤信息而不是 panic

#### Scenario: 超時處理
- **WHEN** API 響應超時
- **THEN** 設定合理的超時時間（例如 5 秒）並返回錯誤而非無限等待
