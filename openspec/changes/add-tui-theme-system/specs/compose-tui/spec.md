## MODIFIED Requirements

### Requirement: 視覺設計與焦點提示

系統 SHALL 由目前選定的 TUI theme 統一控制完整 canvas、panels、輸入元件、狀態列與 overlays 的背景、前景、selection、placeholder、cursor、border 與狀態色；元件不得保留與 theme 無關的固定色碼。焦點 panel MUST 以目前 theme 的 Accent token 清楚標示。

#### Scenario: Header Panel 邊框隨焦點改變樣式

- **WHEN** 焦點在 Header panel
- **THEN** Header panel 只有左側 focus rail 使用目前 theme 的 Accent token，其餘三邊使用 Border token，Composer 四邊皆使用 Border token

#### Scenario: Composer Panel 邊框隨焦點改變樣式

- **WHEN** 焦點在 Composer panel
- **THEN** Composer panel 只有左側 focus rail 使用目前 theme 的 Accent token，其餘三邊使用 Border token，Header 四邊皆使用 Border token

#### Scenario: Preview 不顯示焦點提示

- **WHEN** Compose TUI render Preview panel
- **THEN** Preview 使用直角 border 且四邊皆使用目前 theme 的 Border token

#### Scenario: Border 與 Panel 背景分界一致

- **WHEN** Header、Composer、Preview 或任一 overlay render 直角 border
- **THEN** 四側與四個 corner 的 background 均使用目前 theme 的 Canvas token，border 內部使用 Panel token，不得回落至 terminal default background

#### Scenario: Focus rail 不改變 Border 背景

- **WHEN** Header 或 Composer 取得焦點
- **THEN** 只有左側 border foreground 改為 Accent token，四側 border background 仍全部使用 Canvas token

#### Scenario: 空白 Composer 背景不穿透

- **WHEN** Composer 為空白、部分輸入、focused 或 blurred 狀態
- **THEN** textarea 所有沒有明確 selection background 的可見 cell 均使用目前 theme 的 Panel token，不得顯示 host terminal background

#### Scenario: 輸入元件使用 Theme

- **WHEN** 使用者在 Header 或 Composer 輸入文字
- **THEN** textinput／textarea 的文字、placeholder、cursor、selection 與 focused／blurred styles 均使用目前 theme 的 semantic tokens

#### Scenario: 裝飾性視覺元素

- **WHEN** 撰寫畫面初始化
- **THEN** Header 和 Composer panel 的 border title 可包含 `...` 和 `▽` 作為視覺裝飾（不實現展開/收合），且裝飾色來自目前 theme

#### Scenario: Runtime 切換不改變 Compose State

- **WHEN** 使用者透過 Theme Picker 預覽或確認另一個 theme
- **THEN** Header、Body、Preview、附件、焦點、發信與確認狀態保持不變，只有 style 與目前 theme 改變
