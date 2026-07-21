## Context

`ComposeModel.Update()` 目前直接處理 `Ctrl+C`、`Ctrl+S`、`Esc`、`Ctrl+A`、`Ctrl+J`、`Ctrl+K` 與三組模板快捷鍵；`renderStatusBar()` 也以固定字串列出多個操作。這種做法會讓新增 History、Settings 等功能時持續擴大 key switch 與狀態列，而且現有 `Esc` 同時承擔清除與退出意圖，容易造成破壞性誤操作。

本變更只調整無子命令啟動的 Compose TUI。EML 子命令、Burst mode 與 SMTP 發信介面不在此次範圍內。

## Goals / Non-Goals

**Goals:**

- 提供以 `Ctrl+X` 啟動、可在狀態列探索的 prefix command 模式。
- 集中管理 prefix 指令的 key、label、可用狀態與 semantic command ID。
- 將退出、清除、附件、模板與 Help 移入 prefix namespace。
- 讓 `Esc` 成為一致且安全的取消／返回操作。
- 保留寄送與 panel／欄位導航等高頻直接快捷鍵。
- 以可測試的狀態轉移設計，為後續 History screen command 留下擴充點。

**Non-Goals:**

- 不在此變更新增 History、Settings、Drafts 或其他新畫面。
- 不建立通用 command palette、模糊搜尋或自訂快捷鍵設定。
- 不改變 EML、Burst mode 或 SMTP 寄送行為。
- 不新增 timeout；prefix 模式會等待有效指令或取消。
- 不新增外部 dependency。

## Decisions

### 1. 使用 `Ctrl+X` 作為固定 prefix

`Ctrl+X` 在文字輸入導向的 TUI 中可與一般字元輸入區隔，也不像 Space 會直接輸入內容，且較不易與 tmux 的預設 prefix 衝突。按下後進入明確的 command 狀態，下一個按鍵不會傳給 `textinput` 或 `textarea`。

替代方案包括 Space、Esc、`Ctrl+Space`、`Ctrl+B` 與 `Ctrl+G`；它們分別有文字輸入衝突、取消語意衝突、terminal／IME 相容性、tmux 衝突或傳統 cancel 語意等問題，因此不採用。

### 2. 以獨立 prefix state component 與 semantic command ID 分離按鍵和行為

新增可重用的 prefix state component，保存目前模式，例如：

```text
normal
root-prefix
template-prefix
confirm-clear
confirm-quit
help
```

Command registry 定義 command ID、觸發 key、顯示 label 與適用狀態。Dispatcher 將 key 轉成 semantic command ID，`ComposeModel` 再執行清除、退出、開啟 filepicker 或套用模板等實際行為。這避免 command component 直接依賴 Compose 欄位，也讓後續上層 AppModel 可以接手 `open-history` 等畫面導航指令。

不在本次先引入完整 AppModel screen router，避免為尚未存在的 History screen 重構 AlertModel 與現有 model 切換。

### 3. Command HUD 取代長期顯示所有低頻快捷鍵

正常狀態列只顯示高頻操作與 `[Ctrl+X] Commands`。進入 root prefix 後，狀態列改為顯示：

```text
COMMAND  [A] Attach  [T] Template  [C] Clear  [Q] Quit  [?] Help  [Esc] Cancel
```

進入 template prefix 後顯示 HTML、Plain Text、EML template 的下一層選項。HUD 內容由同一份 command registry 產生，避免顯示文字與實際 key binding 漂移。

### 4. Prefix 不使用 timeout

按下 `Ctrl+X` 後會持續等待下一個指令，直到執行成功、按 Esc 取消或輸入未知 key。這讓互動可預期，也避免速度較慢的使用者被 timeout 中斷，並簡化 headless state transition tests。

未知 key SHALL 取消 prefix、保持 Compose 內容不變，並在狀態列顯示短暫的 unknown command 訊息。

### 5. 破壞性操作使用明確確認狀態

Compose dirty 的判定包含任一 Header 欄位、Body 或附件路徑非空。

- `Ctrl+X`、`c`：若 Compose 為空，結束 prefix 且不做其他變更；若為 dirty，進入 confirm-clear，再按 `c` 清除，Esc 取消。
- `Ctrl+X`、`q`：若 Compose 為空，直接回傳 `tea.Quit`；若為 dirty，進入 confirm-quit，再按 `q` 退出，Esc 取消。

這使單次 Esc、未知指令或誤按 prefix 都不會遺失內容。

### 6. 保留高頻 direct shortcuts，移除舊的低頻 direct bindings

`Ctrl+S`、`Ctrl+J`、`Ctrl+K`、Tab 與 Shift+Tab 維持原行為。附件與模板改由 prefix 進入；`Ctrl+A`、`Ctrl+H`、`Ctrl+T`、`Ctrl+E` 不再觸發原本的應用程式層操作。

退出只透過 prefix quit flow；Compose 不再將 `Ctrl+C` 或連按 Esc 映射為立即退出。這是刻意的互動變更，確保 Quit 會先經過可見且可確認的 command 狀態。

### 7. Key routing 採明確優先順序

訊息處理優先順序為：

```text
Filepicker／Help overlay
        ↓
Confirmation state
        ↓
Prefix dispatcher
        ↓
高頻 global direct shortcuts
        ↓
目前 focus component
```

Overlay 開啟時，Esc 只關閉 overlay；prefix 或其他 Compose 指令不會穿透到背景。Confirmation 與 prefix 狀態中的 key 也不會寫入 Header 或 Body。

## Risks / Trade-offs

- [Risk] 使用者習慣 `Ctrl+C` 或既有附件／模板快捷鍵，遷移後可能一時找不到功能。
  → Mitigation：正常狀態列固定顯示 `[Ctrl+X] Commands`，Command HUD 顯示完整選項，README 同步更新。

- [Risk] 無 timeout 的 prefix 可能讓誤觸後的使用者以為輸入失效。
  → Mitigation：狀態列以明顯的 `COMMAND` 標記目前狀態，Esc 隨時取消，未知 key 也會自動返回 normal。

- [Risk] Command registry 與 Compose action dispatch 增加抽象層。
  → Mitigation：registry 只保存 UI metadata 與 semantic ID，不保存 closure 或業務資料，讓 state tests 保持簡單。

- [Risk] terminal 或既有元件可能對 `Ctrl+X` 有不同處理。
  → Mitigation：由 Compose 的全域 key routing 優先攔截，並以 Bubble Tea v2 `tea.KeyPressMsg` headless tests 驗證 Header、Composer 與 overlay 狀態。

## Migration Plan

1. 先新增 prefix state、command registry 與 state transition tests。
2. 將 Compose 的 Quit、Clear、Attachment、Template、Help 行為接到 semantic command。
3. 更新 Esc、`Ctrl+C` 與舊 direct bindings 的測試，確認不再觸發破壞性操作。
4. 更新 status bar 與 README 快捷鍵說明。
5. 執行 `gofmt`、`go vet ./...` 與專案測試。

若需回滾，可移除 prefix component 並恢復 Compose 原有 direct key switch；此變更不涉及資料格式或持久化 migration。

## Open Questions

無。History command 將由後續獨立 change 定義，屆時可在 registry 加入 `h`，不納入本次 scope。
