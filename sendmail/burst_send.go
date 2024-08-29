package sendmail

import (
	"fmt"
	"hermes/utils"
	"net/mail"
)

// 瘋狂發送郵件
func BurstModeSendMail() {

	from := mail.Address{
		Name: "Hermes", Address: "weiting.shi1982@gmail.com",
	}
	fmt.Printf("==>%s\n", from.String())

	randStr := utils.RandomString(5)
	fmt.Println(".................", randStr)

	emails := make([]string, 100)
	for i := range emails {
		emails[i] = utils.RandomString(5)
	}

	fmt.Println(emails)
}

func GenerateNumberOfEmails(amount int) []string {
	emails := make([]string, amount)
	for i := range emails {
		emails[i] = utils.RandomString(5)
	}
	return emails
}
