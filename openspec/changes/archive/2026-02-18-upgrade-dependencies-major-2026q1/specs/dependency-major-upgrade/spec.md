## ADDED Requirements

### Requirement: 依賴版本升級一致性
Hermes 系統 MUST 將本次定義的核心直接依賴升級至核准版本，並確保 `go.mod` 與 `go.sum` 的相依解析結果一致且可重現。

#### Scenario: 核心依賴完成升級
- **WHEN** 維護者完成本次升級清單中的依賴更新
- **THEN** 專案 MUST 能在乾淨環境中成功解析並下載相依套件

### Requirement: 升級後建置與測試可通過
Hermes 系統 MUST 在依賴升級後維持可建置與可測試狀態，至少通過 `make lint`、`make test` 與 `make build`。

#### Scenario: 升級驗證成功
- **WHEN** 維護者於升級後執行標準驗證指令
- **THEN** 所有指令 MUST 成功完成且不產生阻斷性錯誤

### Requirement: TUI 核心互動行為不可退化
Hermes 系統 MUST 在依賴升級後維持既有 TUI 核心互動能力，包含文字輸入、視窗捲動、檔案挑選與關鍵快捷操作。

#### Scenario: TUI 關鍵流程可正常操作
- **WHEN** 使用者在升級後執行主要 TUI 發信流程
- **THEN** 系統 MUST 保持與升級前等價的可操作性與可預期結果

### Requirement: 不相容變更需有可回復策略
Hermes 系統 MUST 在遇到 major 依賴不相容時提供明確回復路徑，以避免長時間阻斷開發與交付。

#### Scenario: 發生重大不相容
- **WHEN** 升級導致不可接受的功能回歸或阻斷性錯誤
- **THEN** 維護者 MUST 能依既定策略回退問題升級批次並恢復可用狀態
