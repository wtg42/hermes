package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "A command-line tool",
	Long:  `A command-line tool that can perform various operations.`,
}

func init() {
	rootCmd.AddCommand(directSendMailCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error::::", err)
		os.Exit(1)
	}
}
