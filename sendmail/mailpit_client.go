// +build integration

package sendmail

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MailpitMessage 代表 Mailpit API 返回的郵件信息
type MailpitMessage struct {
	ID        string   `json:"ID"`
	From      From     `json:"From"`
	To        []To     `json:"To"`
	Cc        []To     `json:"Cc"`
	Bcc       []To     `json:"Bcc"`
	Subject   string   `json:"Subject"`
	Date      string   `json:"Date"`
	Size      int      `json:"Size"`
	Inline    interface{} `json:"Inline"` // 使用 interface{} 來接受任何類型
	Attachments []Attachment `json:"Attachments"`
	Read      bool     `json:"Read"`
}

// From 代表郵件的發件人
type From struct {
	Name    string `json:"Name"`
	Address string `json:"Address"`
}

// To 代表郵件的收件人
type To struct {
	Name    string `json:"Name"`
	Address string `json:"Address"`
}

// MailpitSearchResponse 代表搜尋結果
type MailpitSearchResponse struct {
	Messages []MailpitMessage `json:"messages"`
	Total    int              `json:"total"`
	Count    int              `json:"count"`
}

const (
	mailpitAPIBase = "http://127.0.0.1:8025/api/v1"
	httpTimeout    = 5 * time.Second
)

// getHTTPClient 返回配置了超時的 HTTP 客戶端
func getHTTPClient() *http.Client {
	return &http.Client{
		Timeout: httpTimeout,
	}
}

// getLatestMessage 從 Mailpit API 獲取最新的郵件
//   - 返回最新郵件的信息，或錯誤
func getLatestMessage() (*MailpitMessage, error) {
	client := getHTTPClient()

	url := fmt.Sprintf("%s/message/latest", mailpitAPIBase)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Mailpit API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Mailpit API returned status %d: %s", resp.StatusCode, string(body))
	}

	var message MailpitMessage
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return nil, fmt.Errorf("failed to parse Mailpit response: %w", err)
	}

	return &message, nil
}

// searchMessages 搜尋符合條件的郵件
//   - query: 搜尋字符串（例如主題、內容等）
//   - 返回匹配的郵件列表或錯誤
func searchMessages(query string) ([]MailpitMessage, error) {
	client := getHTTPClient()

	url := fmt.Sprintf("%s/search?query=%s", mailpitAPIBase, query)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Mailpit API returned status %d: %s", resp.StatusCode, string(body))
	}

	var searchResp MailpitSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	return searchResp.Messages, nil
}

// getRawMessage 獲取郵件的原始格式
//   - id: 郵件 ID（可使用 "latest" 獲取最新郵件）
//   - 返回原始郵件內容或錯誤
func getRawMessage(id string) (string, error) {
	client := getHTTPClient()

	url := fmt.Sprintf("%s/message/%s/raw", mailpitAPIBase, id)
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get raw message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Mailpit API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// getMessageCount 獲取 Mailpit 中的郵件總數
func getMessageCount() (int, error) {
	client := getHTTPClient()

	url := fmt.Sprintf("%s/messages", mailpitAPIBase)
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get message count: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("Mailpit API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Total int `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Total, nil
}
