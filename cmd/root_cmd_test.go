// Commands are added to rootCmd successfully
package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/wtg42/hermes/tui"
)

func TestCommandsAddedToRootCmd(t *testing.T) {
	// 根據 Cobra 預設行為 Commnad 需要按照字母順序加入 否則會出現錯誤
	rootCmd := &cobra.Command{}
	burstModeCmd := &cobra.Command{Use: "burstMode"}
	startTUICmd := &cobra.Command{Use: "startTUI"}

	rootCmd.AddCommand(burstModeCmd)
	rootCmd.AddCommand(startTUICmd)

	assert.Equal(t, 2, len(rootCmd.Commands()))
	assert.Equal(t, "burstMode", rootCmd.Commands()[0].Use)
	assert.Equal(t, "startTUI", rootCmd.Commands()[1].Use)
}

func TestRootCommandThemeSelection(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantTheme string
	}{
		{name: "default gruvbox", wantTheme: "gruvbox"},
		{name: "explicit gruvbox", args: []string{"--theme", "gruvbox"}, wantTheme: "gruvbox"},
		{name: "tokyo night", args: []string{"--theme", "tokyo-night"}, wantTheme: "tokyo-night"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got tui.Theme
			command := newRootCommand(func(theme tui.Theme) error {
				got = theme
				return nil
			})
			command.SetArgs(tt.args)

			if err := command.Execute(); err != nil {
				t.Fatalf("Execute() returned error: %v", err)
			}
			if got.Name != tt.wantTheme {
				t.Fatalf("runner theme = %q, want %q", got.Name, tt.wantTheme)
			}
		})
	}
}

func TestRootCommandRejectsUnknownThemeBeforeRunningTUI(t *testing.T) {
	called := false
	command := newRootCommand(func(theme tui.Theme) error {
		called = true
		return nil
	})
	command.SetArgs([]string{"--theme", "nord"})

	err := command.Execute()
	if err == nil {
		t.Fatal("Execute() unexpectedly accepted unknown theme")
	}
	if called {
		t.Fatal("TUI runner was called for an invalid theme")
	}
	for _, want := range []string{"nord", "gruvbox", "tokyo-night"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not mention %q", err, want)
		}
	}
}
