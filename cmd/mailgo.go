package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mailgoCmd = &cobra.Command{
	Use:   "mailgo",
  Short: "mailgo is a simple mail sender",
  Long: `mailgo is a simple mail sender allowing to send emails to a list of recipients.`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println(":")
  },
}

var subCmd  = &cobra.Command{
  Use: "sender",
  Short: "send is a command to send an email",
  Long: `send is a command to send an email to a list of recipients.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("setting Sender email address")
	},
}

func Execute() {
	mailgoCmd.Execute()
	// subCmd.Execute()
}