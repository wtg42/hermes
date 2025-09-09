# Repository Guidelines

本文件說明在本倉庫貢獻程式碼的基本規範與流程，保持精簡、可維護與可預測。請於提交 PR 前先閱讀。

## Project Structure & Module Organization
- `main.go`：程式進入點；`main_test.go` 為整體行為測試。
- `cmd/`：CLI 子命令（如 `directSendMail`, `burst`, `start-tui`）。
- `tui/`：TUI 介面與互動元件。
- `sendmail/`：郵件發送邏輯（直送、爆發模式）。
- `utils/`：通用輔助函式；`*_test.go` 放同目錄。
- `assets/`：內嵌資源（字體、圖片），由 `assets/assets.go` 讀取。

## Build, Test, and Development Commands
- 建置：`make build`（等同 `go build -o bin/hermes .`）
- 測試：`make test`（等同 `go test ./... -race -cover`）
- 檢查/格式化：`make lint`（等同 `go vet ./... && go fmt ./...`）
- 執行：`make run`（等同 `go run . start-tui`）
- 清理：`make clean`（移除 `bin/`）
- 安裝：`go install` 後可直接執行 `hermes`

## Coding Style & Naming Conventions
- 使用 Go 標準工具（`gofmt`/`goimports`）。避免全域狀態，模組化分層。
- 封包名小寫、無底線；匯出符號採 PascalCase，非匯出採 lowerCamelCase。
- 檔名以底線風格（例如：`mail_burst.go`）與既有慣例一致。
- 每個公開 API 以註解描述用途與錯誤情境（English preferred）。

## Development Approach
- 採用 TDD（Test-Driven Development）方針：先寫失敗測試，再最小實作，最後重構（Red→Green→Refactor）。
- 每次變更先新增/更新測試，確保覆蓋到正向與錯誤路徑；不得以暫時跳過（skip）長期存在。
- PR 必附測試證據：`make test` 輸出摘要或關鍵案例說明。

## Testing Guidelines
- 框架：Go 標準 `testing`；檔名 `*_test.go` 與被測套件同目錄。
- 命名：`TestXxx` 對應公開行為；表格測試優先，涵蓋錯誤路徑。
- 覆蓋率：盡量維持/提升現有覆蓋；新增功能需附測試。

## Commit & Pull Request Guidelines
- Commit 風格：Conventional Commits（例：`feat: ...`, `fix: ...`, `refactor: ...`, `test: ...`）。
- PR 需包含：變更動機、設計簡述、風險/相容性注意、測試證據；若影響 CLI 介面請同步更新 `README.md`。
- 小步提交、單一主題；避免無關格式化；附上關聯 Issue。

## Security & Configuration Tips
- 請勿提交密鑰/憑證/`.env`；設定以參數旗標傳遞（如 `--host`, `--port`）。
- 新增資產放於 `assets/` 並評估體積；避免非必要大型二進位。
