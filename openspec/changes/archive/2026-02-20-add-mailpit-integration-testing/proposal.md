## Why

目前的測試只使用 mock 進行單元測試，缺乏端到端的郵件發送驗證。整合真實的 SMTP 服務器（Mailpit）可以確保郵件實際能夠成功發送，提高對寄信功能的信心。

## What Changes

- 添加 `docker-compose.yml` 在專案根目錄，定義 Mailpit 容器配置
- 修改 `Makefile` 的 `test` target：
  - 檢查 Docker 可用性
  - 啟動 Mailpit 服務
  - 執行測試
  - 關閉 Mailpit 服務
- 添加集成測試用例，實際寄信到 Mailpit 並驗證郵件接收

## Capabilities

### New Capabilities
- `mailpit-integration-testing`: 集成測試框架，使用 Docker Compose 快速啟動 Mailpit，支援自動化端到端郵件發送測試

### Modified Capabilities
- `test-automation`: 修改現有測試流程，整合 Mailpit 服務的啟動和停止管理

## Impact

- 開發工作流：運行 `make test` 時自動管理 Mailpit 生命週期
- CI/CD 友善：可在任何支援 Docker 的環境中運行
- 測試可靠性：真實郵件發送驗證而非純粹 mock
- 依賴項：Docker 和 Docker Compose（需要版本檢查）
