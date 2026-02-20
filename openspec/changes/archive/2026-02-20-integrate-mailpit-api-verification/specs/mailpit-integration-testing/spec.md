# Mailpit Integration Testing

## MODIFIED Requirements

### Requirement: 測試自動化管理 Mailpit 生命週期
`make test` 命令應自動啟動 Mailpit 服務、執行測試和郵件內容驗證，最後關閉服務。

#### Scenario: make test 啟動 Mailpit
- **WHEN** 執行 `make test`
- **THEN** 首先啟動 Mailpit 容器（若尚未運行）

#### Scenario: make test 執行測試套件
- **WHEN** Mailpit 啟動完成後
- **THEN** 執行 `go test ./... -race -cover -tags integration`

#### Scenario: make test 執行郵件內容驗證
- **WHEN** 測試完成後
- **THEN** 測試代碼透過 Mailpit API 驗證郵件內容，確保郵件被正確發送和接收

#### Scenario: make test 清理 Mailpit
- **WHEN** 所有測試和驗證執行完成（成功或失敗）
- **THEN** 執行 `docker-compose down` 關閉並清理容器

#### Scenario: Docker 不可用時提示用戶
- **WHEN** 執行 `make test` 但 Docker 不可用
- **THEN** 顯示錯誤訊息提示安裝 Docker，不執行測試

## ADDED Requirements

### Requirement: Mailpit 容器在測試期間保持運行
Mailpit 容器應在測試執行期間持續運行，以支援多個測試和 API 驗證。

#### Scenario: 容器生命週期管理
- **WHEN** `make test` 執行期間
- **THEN** Mailpit 容器保持運行，直到所有測試和驗證完成

#### Scenario: 多個測試共享同一 Mailpit 實例
- **WHEN** 多個集成測試依序執行
- **THEN** 所有測試都能連接到同一個 Mailpit 實例進行驗證
