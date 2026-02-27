// mail package 定義郵件發信的領域模型與介面
// 提供層與層之間的清晰邊界
package mail

// MailCompose 包含郵件的所有必要信息
// 用於在 TUI、cmd、sendmail 各層之間傳遞郵件數據
type MailCompose struct {
	From       string // 寄件者
	To         []string // 收件者清單
	CC         []string // 副本清單
	BCC        []string // 密件副本清單
	Subject    string // 主旨
	Body       string // 郵件正文
	Attachment string // 附件路徑（空字串表示無附件）
	Host       string // SMTP 伺服器
	Port       string // SMTP 埠號
}

// Mailer 定義郵件發信介面
// 任何實作此介面的型別都可以用於發信
type Mailer interface {
	Send(compose MailCompose) error
}
