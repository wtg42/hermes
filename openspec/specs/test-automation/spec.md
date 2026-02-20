# Test Automation

## Purpose

Define requirements for automated testing workflows including dependency checks, environment cleanup, and proper resource lifecycle management.

## Requirements

### Requirement: 測試命令包含依賴檢查
`make test` 命令應在執行前檢查必要的依賴項（Docker、Docker Compose）。

#### Scenario: Docker 可用時正常執行
- **WHEN** 執行 `make test` 且系統已安裝 Docker 和 Docker Compose
- **THEN** 繼續執行測試流程

#### Scenario: Docker 不可用時中斷並提示
- **WHEN** 執行 `make test` 但 Docker 不可用
- **THEN** 印出詳細的錯誤訊息，建議用戶安裝 Docker，並以非零狀態碼退出

### Requirement: 測試前進行環境清理
`make test` 應確保在啟動 Mailpit 之前沒有舊的容器執行，避免端口衝突。

#### Scenario: 舊容器存在時先清理
- **WHEN** 執行 `make test` 且舊的 Mailpit 容器仍在運行
- **THEN** 先執行 `docker-compose down` 清理，然後啟動新容器

#### Scenario: 第一次執行時正常啟動
- **WHEN** 首次執行 `make test` 且沒有現存容器
- **THEN** 直接啟動 Mailpit 容器

### Requirement: 測試結果不受 Mailpit 狀態影響
測試完成後，Mailpit 容器應被完全移除，不會對後續操作造成干擾。

#### Scenario: 成功測試後清理容器
- **WHEN** 測試執行成功並完成
- **THEN** Mailpit 容器被停止並移除，不留痕跡

#### Scenario: 失敗測試後也進行清理
- **WHEN** 測試執行失敗或超時
- **THEN** 仍然執行清理流程，確保容器被停止並移除
