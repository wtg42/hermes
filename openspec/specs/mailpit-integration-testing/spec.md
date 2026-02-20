# Mailpit Integration Testing

## Purpose

Provide integration testing infrastructure for Hermes using Mailpit as a lightweight SMTP mock server. Enable end-to-end email delivery validation within Docker containers.

## Requirements

### Requirement: Docker Compose 配置提供 Mailpit 服務
系統應提供 `docker-compose.yml` 文件，定義 Mailpit 容器配置，暴露 SMTP 端口 1025 供測試使用。

#### Scenario: docker-compose.yml 存在且配置正確
- **WHEN** 查看專案根目錄
- **THEN** 存在 `docker-compose.yml` 文件，定義 Mailpit 服務，監聽 `127.0.0.1:1025`

#### Scenario: Mailpit 容器可成功啟動
- **WHEN** 執行 `docker-compose up -d`
- **THEN** Mailpit 容器成功啟動，SMTP 服務在 localhost:1025 可用

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

### Requirement: 測試代碼能連接到 Mailpit SMTP 服務
測試代碼應能透過環境變數或預設值連接到 Mailpit 的 SMTP 服務（localhost:1025）。

#### Scenario: 預設連接參數
- **WHEN** 測試代碼執行且無環境變數設定
- **THEN** 使用 `localhost:1025` 作為 SMTP 主機和端口

#### Scenario: 透過環境變數覆蓋連接參數
- **WHEN** 設定 `TEST_SMTP_HOST` 或 `TEST_SMTP_PORT` 環境變數
- **THEN** 測試代碼使用指定的主機和端口而非預設值

### Requirement: 端到端測試驗證郵件發送
至少一個集成測試應實際發送郵件到 Mailpit 並驗證發送成功。

#### Scenario: 郵件成功發送到 Mailpit
- **WHEN** 測試代碼向 localhost:1025 發送郵件
- **THEN** 郵件被 Mailpit 接收，smtp.SendMail 返回 nil（無誤）

#### Scenario: 多個收件者郵件發送
- **WHEN** 發送包含 To、Cc、Bcc 的郵件到 Mailpit
- **THEN** 郵件成功發送，所有收件者都被包含

### Requirement: Mailpit 容器在測試期間保持運行
Mailpit 容器應在測試執行期間持續運行，以支援多個測試和 API 驗證。

#### Scenario: 容器生命週期管理
- **WHEN** `make test` 執行期間
- **THEN** Mailpit 容器保持運行，直到所有測試和驗證完成

#### Scenario: 多個測試共享同一 Mailpit 實例
- **WHEN** 多個集成測試依序執行
- **THEN** 所有測試都能連接到同一個 Mailpit 實例進行驗證
