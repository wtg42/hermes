// Commands are added to rootCmd successfully
package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCommandsAddedToRootCmd(t *testing.T) {
	// 根據 Cobra 預設行為 Commnad 需要按照字母順序加入 否則會出現錯誤
	rootCmd := &cobra.Command{}
	burstModeCmd := &cobra.Command{Use: "burstMode"}
	directSendMailCmd := &cobra.Command{Use: "directSendMail"}
	startTUICmd := &cobra.Command{Use: "startTUI"}

	rootCmd.AddCommand(burstModeCmd)
	rootCmd.AddCommand(directSendMailCmd)
	rootCmd.AddCommand(startTUICmd)

	assert.Equal(t, 3, len(rootCmd.Commands()))
	assert.Equal(t, "burstMode", rootCmd.Commands()[0].Use)
	assert.Equal(t, "directSendMail", rootCmd.Commands()[1].Use)
	assert.Equal(t, "startTUI", rootCmd.Commands()[2].Use)
}
