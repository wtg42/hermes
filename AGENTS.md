<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->


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

## Bubble Tea v2 / TUI Testing Guidelines
- 本專案使用 `charm.land/bubbletea/v2`。測試範例請以 v2 API 為準，鍵盤事件優先使用 `tea.KeyPressMsg`，避免直接照搬 Bubble Tea v1 的 `tea.KeyMsg` 範例。
- Bubble Tea v2 提供適合 headless test 的 program options：`tea.WithInput`, `tea.WithOutput`, `tea.WithWindowSize`, `tea.WithEnvironment`, `tea.WithColorProfile`, `tea.WithoutSignals`；整合測試可用 `Program.Send` 注入 `tea.Msg`，並用 `tea.Quit` 收尾。
- TUI 測試採分層策略：優先測 `Update` 的狀態轉移，其次測固定 model 狀態下的 `View()` 輸出，最後才少量測完整 `tea.Program` 流程。這是本專案測試方針，源自 Bubble Tea v2 API 與其自身測試模式的整理，非官方文件逐字規範。
- 測 `Update` 時直接建立 model 並傳入 message，檢查 focus、shortcut、欄位值、錯誤訊息、sending 狀態等 state machine 行為；避免為了單純狀態邏輯啟動完整 terminal program。
- 測 `View()` 時固定 terminal width/height、環境與色彩能力，必要時比較重要片段而非完整 ANSI escape sequence；只有穩定 layout 才考慮 golden/snapshot。
- 測完整 program 時使用 `bytes.Buffer` 作為 input/output，固定 `tea.WithWindowSize(80, 24)` 與 `tea.WithEnvironment([]string{"TERM=xterm-256color"})`，用 goroutine 呼叫 `p.Send(...)` 注入事件，再送 `tea.Quit()`，避免依賴真實 terminal 或人工互動。
- 新增 TUI 功能時至少覆蓋主要成功路徑與錯誤/取消路徑；若 UI 行為牽涉寄信、副作用或外部依賴，應注入 mock/stub，避免測試真的發信或讀寫使用者環境。

## Commit & Pull Request Guidelines
- Commit 風格：Conventional Commits（例：`feat: ...`, `fix: ...`, `refactor: ...`, `test: ...`）。
- PR 需包含：變更動機、設計簡述、風險/相容性注意、測試證據；若影響 CLI 介面請同步更新 `README.md`。
- 小步提交、單一主題；避免無關格式化；附上關聯 Issue。

## Security & Configuration Tips
- 請勿提交密鑰/憑證/`.env`；設定以參數旗標傳遞（如 `--host`, `--port`）。
- 新增資產放於 `assets/` 並評估體積；避免非必要大型二進位。
