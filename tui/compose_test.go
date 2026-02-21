package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

func TestSplitPaneWidths_OddWidthUsesAllColumns(t *testing.T) {
	left, right := splitPaneWidths(121)

	if left+right != 121 {
		t.Fatalf("expected split widths to consume all columns, got %d+%d", left, right)
	}
	if left != 61 || right != 60 {
		t.Fatalf("expected 61/60 split for odd width, got %d/%d", left, right)
	}
}

func TestRenderHeaderPanel_ShowsFullShortPlaceholders(t *testing.T) {
	m := ComposeModel{
		mailFields: []textinput.Model{
			newHeaderInput("FROM"),
			newHeaderInput("TO"),
			newHeaderInput("CC"),
			newHeaderInput("BCC"),
			newHeaderInput("SUBJECT"),
			newHeaderInput("HOST"),
			newHeaderInput("DEFAULT IS 25"),
		},
	}

	panel := m.renderHeaderPanel(50, 12)

	if !strings.Contains(panel, "TO") {
		t.Fatalf("expected TO placeholder to render completely, got: %q", panel)
	}
	if !strings.Contains(panel, "CC") {
		t.Fatalf("expected CC placeholder to render completely, got: %q", panel)
	}
}

func TestComposePaneDimensions_AreConsistent(t *testing.T) {
	leftWidth, rightWidth := splitPaneWidths(120)
	paneHeight := 22

	m := ComposeModel{
		mailFields: []textinput.Model{
			newHeaderInput("FROM"),
			newHeaderInput("TO"),
			newHeaderInput("CC"),
			newHeaderInput("BCC"),
			newHeaderInput("SUBJECT"),
			newHeaderInput("HOST"),
			newHeaderInput("DEFAULT IS 25"),
		},
		composer: textarea.New(),
		preview:  viewport.New(previewContentWidth(rightWidth), paneHeight),
	}

	m.preview.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1)

	headerHeight, composerHeight := splitLeftPaneHeights(paneHeight)
	header := m.renderHeaderPanel(leftWidth, headerHeight)
	composer := m.renderComposerPanel(leftWidth, composerHeight)
	leftPane := lipgloss.JoinVertical(lipgloss.Left, header, composer)
	rightPane := m.preview.View()

	if got := lipgloss.Width(leftPane); got != leftWidth {
		t.Fatalf("expected left pane width %d, got %d", leftWidth, got)
	}
	if got := lipgloss.Width(rightPane); got != rightWidth {
		t.Fatalf("expected right pane width %d, got %d", rightWidth, got)
	}
	if got := lipgloss.Height(leftPane); got != paneHeight {
		t.Fatalf("expected left pane height %d, got %d", paneHeight, got)
	}
	if got := lipgloss.Height(rightPane); got != paneHeight {
		t.Fatalf("expected right pane height %d, got %d", paneHeight, got)
	}
}

func newHeaderInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 512
	return ti
}

func TestComposeUpdate_SendMailProcessClearsSendingState(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	m := ComposeModel{
		sending: true,
		mailFields: []textinput.Model{
			newHeaderInput("FROM"),
			newHeaderInput("TO"),
			newHeaderInput("CC"),
			newHeaderInput("BCC"),
			newHeaderInput("SUBJECT"),
			newHeaderInput("HOST"),
			newHeaderInput("DEFAULT IS 25"),
		},
		composer: textarea.New(),
		preview:  viewport.New(10, 10),
	}

	updated, _ := m.Update(sendMailProcess{result: false, err: errors.New("smtp down")})

	if _, ok := updated.(AlertModel); !ok {
		t.Fatalf("expected AlertModel after send result, got %T", updated)
	}

	saved, ok := viper.Get("compose-model").(ComposeModel)
	if !ok {
		t.Fatalf("expected compose-model in viper, got %T", viper.Get("compose-model"))
	}

	if saved.sending {
		t.Fatal("expected sending state to be false after send result")
	}
}
