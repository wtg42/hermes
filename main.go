package main

import (
	"go-go-power-mail/cmd"
	"go-go-power-mail/sendmail"
)

func main() {
	cmd.Execute()
	sendmail.DirectSendMail()
}
