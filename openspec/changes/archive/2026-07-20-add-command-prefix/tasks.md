## 1. Prefix state 與 command registry（Red）

- [x] 1.1 新增 Bubble Tea v2 `tea.KeyPressMsg` table tests，覆蓋 `Ctrl+X` 進入 root prefix、Esc 取消、未知 key 返回 normal，以及 command key 不寫入 Header／Composer
- [x] 1.2 新增 command registry tests，驗證 root／template 狀態的 key、label、semantic command ID 與 HUD 顯示使用相同定義
- [x] 1.3 新增無 timeout 的 state test，驗證非 key message 不會使 prefix 自動返回 normal

## 2. Prefix component（Green / Refactor）

- [x] 2.1 新增 prefix state、semantic command ID 與集中式 command registry
- [x] 2.2 實作 dispatcher 的 root prefix、template prefix、Esc 返回與未知指令處理
- [x] 2.3 實作由 registry 產生的 root／template Command HUD 與 unknown command 提示
- [x] 2.4 重構並確認 prefix component 不直接依賴 Compose 欄位或 mailer

## 3. Compose 破壞性操作（Red）

- [x] 3.1 新增 dirty Compose tests，涵蓋任一 Header、Body 或附件非空時 `Ctrl+X c` 進入確認、再次按 `c` 清除及 Esc 取消
- [x] 3.2 新增 Quit tests，涵蓋空白 Compose 的 `Ctrl+X q` 直接退出、dirty Compose 二次 `q` 確認及 Esc 取消
- [x] 3.3 新增安全 Esc／Ctrl+C tests，驗證 normal 狀態下不清除內容也不立即退出

## 4. Compose command 整合（Green / Refactor）

- [x] 4.1 將 prefix component 與 dirty-state helper 整合進 `ComposeModel`
- [x] 4.2 實作 Clear 與 Quit confirmation state，確保清除時同步重設 Header、Body、Preview 與附件
- [x] 4.3 依 overlay、confirmation、prefix、direct shortcut、focused component 的優先序重構 key routing
- [x] 4.4 移除 Esc 清除／退出、Ctrl+C 立即退出，以及 Ctrl+A／Ctrl+H／Ctrl+T／Ctrl+E 的舊應用程式層 direct bindings

## 5. Attach、Template 與 Help（Red → Green）

- [x] 5.1 新增 `Ctrl+X a` 開啟 Filepicker 及 Filepicker Esc 僅取消 overlay 的 tests，並完成 Attach command 整合
- [x] 5.2 新增 template prefix 的 HTML／Plain Text／EML 選擇、Esc 返回 root prefix 與 Preview 同步 tests，並完成 Template command 整合
- [x] 5.3 新增 `Ctrl+X ?` Help overlay、Esc 關閉及背景按鍵不穿透 tests，並完成 Help command 整合

## 6. 狀態列與既有行為回歸

- [x] 6.1 新增 View tests，驗證 normal 狀態顯示 `[Ctrl+S] Send`、`[Ctrl+X] Commands` 與 panel hint
- [x] 6.2 新增 View tests，驗證 root prefix、template prefix、Clear／Quit confirmation 與 unknown command 的狀態列內容
- [x] 6.3 驗證 `Ctrl+S`、`Ctrl+J`、`Ctrl+K`、Tab、Shift+Tab、SMTP 狀態及 sending 狀態列維持既有行為

## 7. 文件與驗證

- [x] 7.1 更新 README 的 Compose 快捷鍵與 Command HUD 操作說明，標示舊快捷鍵遷移方式
- [x] 7.2 執行 `gofmt`、`go vet ./...` 與相關 TUI unit tests，修正所有失敗
- [x] 7.3 執行 `make test` 完成 race、coverage 與 integration 驗證，保存測試輸出摘要
