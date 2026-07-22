## MODIFIED Requirements

### Requirement: 視覺設計與焦點提示

系統 SHALL 由目前選定的 TUI theme 統一控制完整 canvas、panels、輸入元件、狀態列與 overlays 的背景、前景、selection、placeholder、cursor、border 與狀態色；元件不得保留與 theme 無關的固定色碼。焦點 panel MUST 以目前 theme 的 Accent token 清楚標示。

#### Scenario: Header Panel 邊框隨焦點改變樣式

- **WHEN** 焦點在 Header panel
- **THEN** Header panel 邊框使用目前 theme 的 Accent token，Composer 邊框使用 Border token

#### Scenario: Composer Panel 邊框隨焦點改變樣式

- **WHEN** 焦點在 Composer panel
- **THEN** Composer panel 邊框使用目前 theme 的 Accent token，Header panel 邊框使用 Border token

#### Scenario: 輸入元件使用 Theme

- **WHEN** 使用者在 Header 或 Composer 輸入文字
- **THEN** textinput／textarea 的文字、placeholder、cursor、selection 與 focused／blurred styles 均使用目前 theme 的 semantic tokens

#### Scenario: 裝飾性視覺元素

- **WHEN** 撰寫畫面初始化
- **THEN** Header 和 Composer panel 的 border title 可包含 `...` 和 `▽` 作為視覺裝飾（不實現展開/收合），且裝飾色來自目前 theme

#### Scenario: Runtime 切換不改變 Compose State

- **WHEN** 使用者透過 Theme Picker 預覽或確認另一個 theme
- **THEN** Header、Body、Preview、附件、焦點、發信與確認狀態保持不變，只有 style 與目前 theme 改變
