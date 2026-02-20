## Context

目前 Makefile 的 `test` target 直接執行 `go test ./... -race -cover`。專案沒有集成測試基礎設施，發信邏輯使用 mock 驗證（注入假的 SendMail 函數）。

需要整合 Mailpit（輕量級 SMTP 服務器）以支援真實端到端郵件發送測試，同時保持現有 mock 測試的獨立性。

## Goals / Non-Goals

**Goals:**
- 提供簡單的 Docker Compose 配置來啟動 Mailpit
- 修改 `make test` 自動管理 Mailpit 生命週期（啟動 → 測試 → 停止）
- 添加環境變數方式讓測試代碼可配置 SMTP 主機和端口
- 優雅的失敗處理（Docker 不可用時提示，而非中斷）

**Non-Goals:**
- 添加郵件內容驗證（檢查郵件正文、附件等細節）
- 集成 Mailpit Web UI 或 API 查詢郵件
- 支援多個郵件服務器同時運行

## Decisions

### 1. 使用 Docker Compose 而非單獨安裝 Mailpit

**決策**：在 `docker-compose.yml` 中定義 Mailpit 容器，透過 Docker Compose 命令管理生命週期

**理由**：
- 環境一致性：不依賴本地安裝的 Mailpit 版本
- 易於清理：容器停止即自動清理網路和卷
- 易於部署：CI/CD 環境通常已有 Docker

**考慮的替代方案**：
- 手動安裝 Mailpit 二進制：需要開發者本地安裝，維護複雜
- 使用 MailHog：功能相同，但 Mailpit 更新、社群更活躍

### 2. Mailpit 容器暴露特定端口

**決策**：Mailpit SMTP 監聽 `127.0.0.1:1025`（localhost），HTTP UI 監聽 `127.0.0.1:8025`

**理由**：
- 避免與本地郵件服務競爭
- 測試代碼可硬編碼或透過環境變數指定
- 1025 是常見的測試用 SMTP 端口

### 3. Makefile 中的 Docker 檢查和容器管理

**決策**：在 `test` target 中：
1. 檢查 Docker 可用性（`docker ps`）
2. 啟動容器：`docker-compose up -d`
3. 等待 Mailpit 就緒（簡單的延遲或健康檢查）
4. 執行測試：`go test ./...`
5. 清理：`docker-compose down`

**理由**：
- 自動化整個流程，開發者只需 `make test`
- 容器隔離不污染測試環境
- `down` 確保每次測試開始前狀態乾淨

### 4. 環境變數方式配置測試連線

**決策**：測試代碼透過環境變數 `TEST_SMTP_HOST` 和 `TEST_SMTP_PORT` 讀取，預設值 `localhost:1025`

**理由**：
- CI/CD 可輕易覆蓋連線參數
- 本地開發時可忽略，使用預設值
- 不修改 viper 配置邏輯

## Risks / Trade-offs

| 風險 | 減緩策略 |
|-----|--------|
| Docker 不可用或版本過舊 | 檢查 `docker version`，若失敗則提示用戶並中斷，建議安裝 Docker |
| Mailpit 容器啟動緩慢 | 添加簡單延遲（如 sleep 2）或在容器配置中添加健康檢查 |
| 舊 docker-compose 版本不支援某些語法 | 使用簡單配置，僅需最基本功能（image、ports、volumes） |
| 多進程並行執行 make test 可能衝突 | 預期開發工作流是單一進程，CI/CD 環境無此問題 |

## Migration Plan

1. 添加 `docker-compose.yml` 到根目錄
2. 修改 `Makefile` 的 `test` target
3. 添加環境變數文件說明（`.env.test.example` 或在 Makefile 註釋中）
4. 添加簡單集成測試（發信到 localhost:1025 並驗證無誤）
5. 更新 README（可選）說明如何執行測試

**回滾策略**：若 Docker 相關代碼有問題，可直接編輯 Makefile 移除 docker-compose 相關行，回到原始 `go test ./...`

## Open Questions

- 健康檢查：是否需要在 docker-compose.yml 中定義？（或簡單的 sleep 就夠）
- 容器日誌：若測試失敗，是否需要在停止時列印 Mailpit 日誌？
