## 1. 環境設置與準備

- [x] 1.1 驗證 gopls 已安裝並取得版本信息
- [x] 1.2 檢查現有 skill 目錄結構，備份原始文件
- [x] 1.3 建立開發環境的 Go 項目作為測試用例
- [x] 1.4 確認 $GOROOT 和 $GOPATH 環境變量正確配置

## 2. LSP JSON-RPC 通訊層

- [x] 2.1 建立 `lsp-query.sh` 文件，實現基本的 JSON-RPC 客戶端
- [x] 2.2 實現 gopls 進程管理（啟動、停止、狀態檢查）
- [x] 2.3 實現 `textDocument/hover` 請求邏輯
- [x] 2.4 實現 `textDocument/definition` 請求邏輯
- [x] 2.5 實現 `textDocument/references` 請求邏輯
- [x] 2.6 實現 LSP 初始化握手協議 (initialize/initialized)
- [x] 2.7 加入 gopls 不可用時的錯誤處理和日誌

## 3. 符號信息提取與格式化

- [x] 3.1 實現從 LSP hover 回應中提取簽名的邏輯
- [x] 3.2 實現文檔字符串提取和清理（去除 Markdown）
- [x] 3.3 實現型別信息提取
- [x] 3.4 實現已廢棄 API 檢測邏輯
- [x] 3.5 實現 LLM 友好的 Markdown 格式化

## 4. 快取機制

- [x] 4.1 在 `lsp-query.sh` 中實現基於 (file:line:col) 的快取鍵生成
- [x] 4.2 實現快取讀取邏輯（檢查 TTL）
- [x] 4.3 實現快取寫入邏輯
- [x] 4.4 實現快取過期清理（15 分鐘 TTL）
- [x] 4.5 測試快取命中和過期情況

## 5. 集成至現有系統

- [x] 5.1 修改 `trigger-logic.sh`，用 `lsp-query.sh` 替換正則解析
- [x] 5.2 修改 `llm-integration.sh`，新增 `--lsp-query <file> <line> <col>` 選項
- [x] 5.3 保留 `go-source-lookup.sh` 的向後相容性（降級方案）
- [x] 5.4 更新 `format-for-llm.sh` 支持新的 LSP 結果格式
- [x] 5.5 驗證舊接口 `llm-integration.sh --lsp` 仍可工作

## 6. 測試與驗證

- [x] 6.1 為 `lsp-query.sh` 編寫單元測試（gopls 通訊）
- [x] 6.2 測試標準庫符號查詢（例如 `fmt.Println`）
- [x] 6.3 測試依賴包符號查詢
- [x] 6.4 測試已廢棄 API 檢測
- [x] 6.5 測試 LSP 服務不可用時的降級行為
- [x] 6.6 測試快取命中和失效
- [x] 6.7 性能測試：確保不會因 LSP 而降低查詢速度
- [x] 6.8 回歸測試：對比舊實現和新實現的結果

## 7. 文檔與發佈

- [x] 7.1 更新 README.md，說明新的 LSP 查詢接口
- [x] 7.2 更新 DEVELOPMENT.md，記錄 LSP 架構變更
- [x] 7.3 添加故障排查指南（gopls 問題診斷）
- [x] 7.4 更新 SKILL.md，更新功能描述
- [x] 7.5 清理測試文件和臨時日誌
