## Why

目前 hermes 的主流程分成兩個獨立畫面（Header 欄位填寫 → 內文撰寫），使用者體驗割裂，無法同時看到所有欄位與內文。新設計將兩個畫面合併為單一撰寫頁面，並加入即時預覽，提升撰寫體驗。

## What Changes

- 新增 `ComposeModel`：單一畫面整合 Header 欄位、Composer 內文、Preview 預覽
- 左側分割為 Header panel（上）和 Composer panel（下），右側為 Preview panel（永遠顯示）
- 以 `Ctrl+J` / `Ctrl+K` 在 Header panel 和 Composer panel 之間切換焦點
- 底部顯示狀態列，包含快捷鍵提示與 SMTP 連線狀態
- 附件選取改為 Filepicker Overlay 方式，覆蓋在 Composer 區域上方
- 移除舊的兩步驟流程（`MailFieldsModel` → `MailMsgModel`）**BREAKING**

## Capabilities

### New Capabilities

- `compose-tui`: 統一撰寫頁面 TUI，整合 Header 欄位輸入、Composer 多行內文、右側 Preview 即時同步、底部狀態列與 Filepicker Overlay 附件選取

### Modified Capabilities

（無）

## Impact

- `tui/` 目錄：新增 `compose.go`，舊有 `mail_field.go` 和 `mail_msg_contents.go` 保留但不再作為主流程入口
- `cmd/root_cmd.go`：改為啟動 `InitialComposeModel()`
- 不影響 `sendmail/` 發信邏輯與 `viper` 資料傳遞機制
- 不影響 `burst` 和 `eml` 子命令
- 無新增外部依賴（glamour Markdown 渲染為未來計劃）
