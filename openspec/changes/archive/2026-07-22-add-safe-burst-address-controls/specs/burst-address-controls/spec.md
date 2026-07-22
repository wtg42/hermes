## ADDED Requirements

### Requirement: Burst 地址可分別固定或隨機產生
系統 SHALL 允許 Burst mode 的 From 與 To 各自使用單一固定 mailbox address，或在未指定固定值時為每封郵件產生隨機地址。固定值 SHALL 同時套用至 MIME header 與 SMTP envelope。

#### Scenario: From 與 To 都固定
- **WHEN** 使用者提供有效的 `--from` 與 `--to`，且兩者網域皆已獲授權
- **THEN** 系統在整批郵件的 header 與 SMTP envelope 使用指定 From 與 To，不要求 `--domain`

#### Scenario: From 固定而 To 隨機
- **WHEN** 使用者提供有效的 `--from`，未提供 `--to`，並提供已獲授權的 `--domain`
- **THEN** 系統在整批郵件使用固定 From，並為每封郵件從指定網域池產生 To

#### Scenario: From 隨機而 To 固定
- **WHEN** 使用者未提供 `--from`，提供有效的 `--to`，並提供已獲授權的 `--domain`
- **THEN** 系統為每封郵件從指定網域池產生 From，並在整批郵件使用固定 To

#### Scenario: From 與 To 都隨機
- **WHEN** 使用者未提供 `--from` 與 `--to`，並提供已獲授權的 `--domain`
- **THEN** 系統為每封郵件的 From 與 To 分別從指定網域池產生隨機地址

### Requirement: 隨機地址必須使用明確網域
只要 From 或 To 任一方需要隨機產生，系統 MUST 要求至少一個有效的 `--domain`，且 SHALL NOT 從 SMTP host 或內建值推導預設網域。

#### Scenario: 隨機欄位缺少 domain
- **WHEN** From 或 To 任一方未指定，且使用者未提供 `--domain`
- **THEN** 系統在發送前回傳缺少隨機地址網域的錯誤，不送出任何郵件

#### Scenario: 多個隨機網域
- **WHEN** 使用者提供多個格式有效且已獲授權的 `--domain` 值
- **THEN** 系統使用這些網域建立隨機地址池，供所有未指定的 From 或 To 抽選

### Requirement: Burst 網域採正向表列與精確授權
系統 SHALL 內建 `rd01.softnext.com.tw` 作為安全網域。每個實際用於固定 From、固定 To 或隨機地址的其他網域 MUST 精確出現在本次執行的 `--allow-domain` 中；父網域授權 SHALL NOT 自動授權其子網域。

#### Scenario: 使用內建安全網域
- **WHEN** 本次執行使用的所有地址網域皆為 `rd01.softnext.com.tw`
- **THEN** 系統不要求額外的 `--allow-domain`

#### Scenario: 非表列網域未授權
- **WHEN** 固定地址或 `--domain` 使用 `softnext.com.tw`，但未提供相符的 `--allow-domain softnext.com.tw`
- **THEN** 系統列出未授權網域並拒絕整批發送

#### Scenario: 非表列網域獲得精確授權
- **WHEN** 固定地址或 `--domain` 使用 `softnext.com.tw`，且使用者提供 `--allow-domain softnext.com.tw`
- **THEN** 系統允許該網域通過安全驗證

#### Scenario: 固定公開網域收件人需要授權
- **WHEN** 使用者提供 Gmail 固定收件人，但未提供 `--allow-domain gmail.com`
- **THEN** 系統拒絕整批發送並指出 `gmail.com` 未獲授權

#### Scenario: 父網域不授權子網域
- **WHEN** 使用者授權 `softnext.com.tw`，但實際地址使用 `mail.softnext.com.tw`
- **THEN** 系統仍將 `mail.softnext.com.tw` 視為未授權網域

#### Scenario: 多網域必須全部通過
- **WHEN** 本次執行使用多個網域，其中至少一個既不在內建表列也未由 `--allow-domain` 精確授權
- **THEN** 系統拒絕整批發送並列出所有未授權網域

### Requirement: Burst 必須在任何發送前完成驗證
系統 MUST 在建立寄信 goroutine 或呼叫 SMTP 前，驗證所有固定地址、隨機網域、授權網域及完整授權關係。任一驗證失敗時 SHALL 回傳錯誤且不得送出部分郵件。

#### Scenario: 固定地址格式錯誤
- **WHEN** `--from` 或 `--to` 不是有效的單一 mailbox address
- **THEN** 系統回傳指出欄位與無效值的錯誤，SMTP 發送次數為零

#### Scenario: 網域格式錯誤
- **WHEN** `--domain` 或 `--allow-domain` 包含格式無效的網域
- **THEN** 系統回傳指出無效網域的錯誤，SMTP 發送次數為零

#### Scenario: 通過預先驗證
- **WHEN** 所有必要地址、網域及授權皆有效
- **THEN** 系統才依 quantity 啟動併發發送，並保留隨機主旨與內容
