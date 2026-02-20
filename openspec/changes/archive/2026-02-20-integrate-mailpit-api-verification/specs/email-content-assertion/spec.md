# 郵件內容斷言

## Purpose

在集成測試中驗證郵件的詳細內容，包括主題、正文、收件人、附件等，確保郵件發送功能的完整性。

## Requirements

### Requirement: 驗證郵件基本信息
測試應能驗證郵件的主題、發件人和收件人是否正確。

#### Scenario: 驗證郵件主題
- **WHEN** 發送包含特定主題的郵件
- **THEN** 通過 API 查詢後，主題應與發送時的設定完全相同

#### Scenario: 驗證發件人
- **WHEN** 從特定的發件者地址發送郵件
- **THEN** 郵件的 From 字段應與設定相符

#### Scenario: 驗證收件人列表
- **WHEN** 發送郵件給多個 To、Cc、Bcc 收件人
- **THEN** 郵件頭應包含所有 To 和 Cc 收件人（BCC 不應出現在郵件頭）

### Requirement: 驗證郵件正文內容
測試應能驗證郵件的文字和 HTML 內容是否正確傳遞。

#### Scenario: 驗證純文字正文
- **WHEN** 發送包含中文和特殊字符的純文字郵件
- **THEN** 郵件的文本部分應能正確解碼，內容應與發送時相同

#### Scenario: 驗證 HTML 內容
- **WHEN** 郵件包含 HTML 部分
- **THEN** 郵件應包含 text/html 類型的 MIME 部分

### Requirement: 驗證字符編碼
確保中文和其他 UTF-8 字符被正確編碼和解碼。

#### Scenario: 主題中文編碼
- **WHEN** 郵件主題包含中文
- **THEN** API 返回的主題應能正確顯示中文，不應出現亂碼

#### Scenario: 正文中文編碼
- **WHEN** 郵件正文包含中文內容
- **THEN** 郵件內容應被正確編碼為 base64，解碼後應正確顯示中文

### Requirement: 驗證附件
測試應能驗證附件是否被正確包含在郵件中。

#### Scenario: 驗證附件存在
- **WHEN** 發送包含附件的郵件
- **THEN** 郵件應包含 multipart/mixed 類型的部分，且包含 Content-Disposition: attachment

#### Scenario: 驗證附件文件名
- **WHEN** 郵件包含附件
- **THEN** 附件的 Content-Disposition 應包含正確的 filename 參數

### Requirement: 驗證 MIME 結構
確保郵件的 MIME 結構符合 RFC 標準。

#### Scenario: 正確的 MIME 邊界
- **WHEN** 郵件為 multipart 格式
- **THEN** 郵件應包含正確的 boundary 分隔符，分隔各個 MIME 部分

#### Scenario: Content-Type 正確性
- **WHEN** 驗證郵件結構
- **THEN** 各個 MIME 部分的 Content-Type 應正確標示（text/plain、text/html、application/* 等）
