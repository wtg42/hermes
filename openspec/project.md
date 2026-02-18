# Project Context

## Purpose
Hermes 是一個以 Go 撰寫的 CLI/TUI 郵件發送工具，目標是讓使用者能快速透過 SMTP 進行單次發信、互動式編輯發信，以及高併發壓力型發信（burst mode）。

核心目標：
- 提供簡潔且可預測的命令列介面（CLI）
- 提供可操作且可測試的文字介面（TUI）
- 保持發信流程模組化，便於擴充與維護

## Tech Stack
- 語言：Go（go.mod: `go 1.23.0`，toolchain `go1.24.2`）
- CLI 框架：`spf13/cobra`
- 設定管理：`spf13/viper`
- TUI 生態：`charmbracelet/bubbletea`, `bubbles`, `lipgloss`
- 測試：Go `testing` + `stretchr/testify`
- 建置與開發：`make`, `go build`, `go test`, `go vet`, `go fmt`

## Project Conventions

### Code Style
- 遵循 Go 標準格式化與靜態檢查：`gofmt`/`go fmt`、`go vet`
- 套件命名使用小寫且不含底線；匯出符號使用 PascalCase，非匯出符號使用 lowerCamelCase
- 檔名維持既有慣例（多使用底線風格，例如 `mail_burst.go`）
- 避免全域狀態，偏好可注入依賴與明確輸入/輸出
- 公開 API 建議附註解，描述用途與錯誤情境（English preferred）

### Architecture Patterns
- 入口為 `main.go`，命令分層於 `cmd/`，業務邏輯主要在 `sendmail/`，互動界面在 `tui/`
- 採分層設計：命令解析（`cmd/`）與郵件發送邏輯（`sendmail/`）分離
- `utils/` 放置可重用輔助函式；`assets/` 管理嵌入資源並由 `assets/assets.go` 統一存取
- 優先使用小型、單一職責模組，避免過早抽象

### Testing Strategy
- 採 TDD（Red -> Green -> Refactor）
- 每次功能變更需新增或更新對應測試，覆蓋成功與錯誤路徑
- 測試檔與被測檔同目錄，命名 `*_test.go`，公開行為使用 `TestXxx`
- PR 前至少執行：`make test`（`go test ./... -race -cover`）

### Git Workflow
- 採小步提交與單一主題變更，避免混入無關格式調整
- Commit 訊息採 Conventional Commits（如 `feat:`, `fix:`, `refactor:`, `test:`）
- 透過 OpenSpec 管理重大功能變更：先提 proposal，審核通過後再實作
- 影響 CLI 介面時，需同步更新 `README.md`

## Domain Context
- 產品領域為 SMTP 郵件發送工具，支援一般發信與測試用途的高併發 burst 發信
- 支援 `to`/`cc`/`bcc` 多收件人（逗號分隔），並進行地址有效性處理
- TUI 模式提供互動式編輯與快捷鍵，適合需要快速調整內容的操作場景
- 專案同時重視可操作性（CLI/TUI）與可測試性（命令與邏輯分層）

## Important Constraints
- 禁止提交敏感資訊（例如 `.env`、密鑰、憑證）
- 以參數旗標傳遞執行設定（如 `--host`, `--port`），避免硬編碼環境資訊
- 新增功能需保持與既有 CLI 命令風格一致，避免破壞既有使用方式
- 新增資產需放在 `assets/` 並注意檔案體積

## External Dependencies
- 外部服務：SMTP 伺服器（例如 Gmail SMTP 或組織內部 MTA）
- Go 第三方套件：Cobra/Viper、Charmbracelet TUI、Testify
- 本機工具鏈：Go toolchain、Make
