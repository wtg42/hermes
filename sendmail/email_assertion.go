//go:build integration
// +build integration

package sendmail

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

// AssertEmailContent 包含郵件斷言的輔助函數

// assertSubjectEquals 驗證郵件主題是否與預期相符
//   - t: 測試上下文
//   - message: Mailpit API 返回的郵件信息
//   - expected: 預期的主題
func assertSubjectEquals(t *testing.T, message *MailpitMessage, expected string) {
	// 主題可能被編碼為 RFC2047 格式，需要解碼
	actualSubject := decodeRFC2047(message.Subject)
	if actualSubject != expected {
		t.Errorf("Subject mismatch: expected %q, got %q", expected, actualSubject)
	}
}

// assertFromEquals 驗證發件人是否與預期相符
//   - t: 測試上下文
//   - message: Mailpit API 返回的郵件信息
//   - expectedAddress: 預期的發件人地址
func assertFromEquals(t *testing.T, message *MailpitMessage, expectedAddress string) {
	if message.From.Address != expectedAddress {
		t.Errorf("From address mismatch: expected %q, got %q", expectedAddress, message.From.Address)
	}
}

// assertToContains 驗證收件人列表是否包含預期的地址
//   - t: 測試上下文
//   - message: Mailpit API 返回的郵件信息
//   - expectedAddresses: 預期的收件人地址列表
func assertToContains(t *testing.T, message *MailpitMessage, expectedAddresses []string) {
	for _, expected := range expectedAddresses {
		found := false
		for _, to := range message.To {
			if to.Address == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("To address not found: %q", expected)
		}
	}
}

// assertCcContains 驗證 Cc 列表是否包含預期的地址
//   - t: 測試上下文
//   - message: Mailpit API 返回的郵件信息
//   - expectedAddresses: 預期的 Cc 地址列表
func assertCcContains(t *testing.T, message *MailpitMessage, expectedAddresses []string) {
	for _, expected := range expectedAddresses {
		found := false
		for _, cc := range message.Cc {
			if cc.Address == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Cc address not found: %q", expected)
		}
	}
}

// assertBccNotInHeaders 驗證 BCC 不出現在郵件頭中
//   - t: 測試上下文
//   - rawMessage: 郵件的原始格式
//   - bccAddress: BCC 地址（應該不在郵件頭中）
func assertBccNotInHeaders(t *testing.T, rawMessage string, bccAddress string) {
	// 分離郵件頭和正文
	headerEnd := strings.Index(rawMessage, "\r\n\r\n")
	if headerEnd == -1 {
		headerEnd = strings.Index(rawMessage, "\n\n")
		if headerEnd == -1 {
			t.Fatal("Could not find message headers end")
		}
	}

	headers := rawMessage[:headerEnd]

	// 只檢查 BCC header 行，不檢查正文中是否包含 BCC 地址
	lines := strings.Split(headers, "\n")
	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		if strings.HasPrefix(lowerLine, "bcc:") {
			if strings.Contains(lowerLine, strings.ToLower(bccAddress)) {
				t.Errorf("BCC address should not appear in BCC header: %q", bccAddress)
			}
		}
	}
}

// assertContentContains 驗證郵件正文是否包含預期的內容
//   - t: 測試上下文
//   - rawMessage: 郵件的原始格式
//   - expectedContent: 預期的內容
func assertContentContains(t *testing.T, rawMessage string, expectedContent string) {
	// 分離郵件頭和正文
	headerEnd := strings.Index(rawMessage, "\r\n\r\n")
	if headerEnd == -1 {
		headerEnd = strings.Index(rawMessage, "\n\n")
		if headerEnd == -1 {
			t.Fatal("Could not find message headers end")
		}
	}

	body := rawMessage[headerEnd:]

	// 內容可能被 base64 編碼，嘗試解碼
	if strings.Contains(body, "Content-Transfer-Encoding: base64") {
		// 嘗試從 base64 編碼的部分提取內容
		encodedContent := extractBase64Content(body)
		if encodedContent != "" {
			decoded, err := base64.StdEncoding.DecodeString(encodedContent)
			if err == nil {
				body = string(decoded)
			}
		}
	}

	if !strings.Contains(body, expectedContent) {
		t.Errorf("Expected content not found in message body: %q", expectedContent)
	}
}

// assertAttachmentExists 驗證郵件中是否存在特定的附件
//   - t: 測試上下文
//   - message: Mailpit API 返回的郵件信息
//   - expectedFileName: 預期的附件文件名
func assertAttachmentExists(t *testing.T, message *MailpitMessage, expectedFileName string) {
	found := false
	for _, att := range message.Attachments {
		if att.FileName == expectedFileName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Attachment not found: %q", expectedFileName)
	}
}

// assertMIMEStructure 驗證郵件的 MIME 結構
//   - t: 測試上下文
//   - rawMessage: 郵件的原始格式
//   - expectedContentType: 預期的 Content-Type（例如 multipart/mixed）
func assertMIMEStructure(t *testing.T, rawMessage string, expectedContentType string) {
	headerEnd := strings.Index(rawMessage, "\r\n\r\n")
	if headerEnd == -1 {
		headerEnd = strings.Index(rawMessage, "\n\n")
	}

	headers := rawMessage[:headerEnd]

	if !strings.Contains(headers, fmt.Sprintf("Content-Type: %s", expectedContentType)) {
		if !strings.Contains(headers, expectedContentType) {
			t.Errorf("Expected Content-Type %q not found in headers", expectedContentType)
		}
	}
}

// decodeRFC2047 解碼 RFC2047 編碼的字符串
func decodeRFC2047(encoded string) string {
	// RFC2047 格式: =?UTF-8?B?...?= (base64) 或 =?UTF-8?Q?...?= (quoted-printable)
	if !strings.HasPrefix(encoded, "=?UTF-8?B?") {
		return encoded
	}

	// 移除前綴和後綴
	content := strings.TrimPrefix(encoded, "=?UTF-8?B?")
	content = strings.TrimSuffix(content, "?=")

	// 解碼 base64
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return encoded
	}

	return string(decoded)
}

// extractBase64Content 從 MIME 正文中提取 base64 編碼的內容
func extractBase64Content(body string) string {
	// 尋找 Content-Transfer-Encoding: base64 後面的內容
	parts := strings.Split(body, "\r\n")
	var content string

	for i, part := range parts {
		if strings.Contains(part, "Content-Transfer-Encoding: base64") {
			// 找到內容的開始位置（在空行之後）
			for j := i + 1; j < len(parts); j++ {
				if parts[j] == "" {
					// 空行表示内容開始
					if j+1 < len(parts) {
						// 收集所有非空行直到下一個邊界
						for k := j + 1; k < len(parts); k++ {
							if strings.HasPrefix(parts[k], "--") {
								// 遇到邊界標記，停止
								break
							}
							if parts[k] != "" {
								content += parts[k]
							}
						}
					}
					break
				}
			}
			break
		}
	}

	return content
}
