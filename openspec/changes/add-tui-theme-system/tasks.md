## 1. Theme Registry（Red → Green）

- [x] 1.1 新增 theme registry tests，驗證唯一名稱、穩定順序、Gruvbox default、Tokyo Night resolution 與未知名稱錯誤
- [x] 1.2 新增 semantic token completeness tests 與 Gruvbox／Tokyo Night 代表色斷言
- [x] 1.3 實作 immutable Theme model、兩個內建 palettes、resolver 與 available-theme API

## 2. Bubble Components Theme 套用（Red → Green）

- [x] 2.1 新增 Compose component style tests，涵蓋 canvas、panels、textinput、textarea、placeholder、selection、Preview 與狀態色
- [x] 2.2 實作集中式 `applyTheme`，同步更新 ComposeModel 與 Bubbles v2 子元件 styles
- [x] 2.3 將 `components.go`、Compose panels、status、Help、Filepicker 與安全確認 overlay 的裸色碼替換為 semantic tokens
- [x] 2.4 新增 Gruvbox／Tokyo Night View tests，驗證完整 canvas 填色且 theme 切換不改變 Compose state

## 3. CLI Theme 選擇（Red → Green）

- [x] 3.1 新增 root command tests，驗證未指定時使用 Gruvbox、`--theme tokyo-night` 與未知名稱 fail closed
- [x] 3.2 實作 root `--theme` flag 與 TUI 初始化注入，保留既有 `InitialComposeModel` 的 Gruvbox default 行為

## 4. Prefix Palette 與 Theme Picker（Red → Green）

- [x] 4.1 擴充 command registry tests，驗證 `Ctrl+X`、`p` 的 Palette semantic command、HUD 與 Help 來源一致
- [x] 4.2 新增 Theme Picker state tests，涵蓋開啟、上下／j／k 預覽、Enter 確認、Esc rollback 與按鍵隔離
- [x] 4.3 實作 Theme Picker overlay 與 runtime `applyTheme`，確保草稿、附件、焦點及確認狀態不變

## 5. 文件與驗證

- [x] 5.1 更新 README，說明預設 Gruvbox、`--theme` 與 `Ctrl+X`、`p` runtime 選擇
- [x] 5.2 執行 raw-color source guard、`gofmt`、`go vet ./...` 與 TUI／CLI unit tests
- [x] 5.3 執行 `make test` 完成 race、coverage 與 Mailpit integration 驗證
- [x] 5.4 執行 OpenSpec strict validation 並確認所有 tasks 完成
