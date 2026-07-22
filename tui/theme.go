package tui

import "fmt"

// DefaultThemeName is the theme used when no explicit selection is provided.
const DefaultThemeName = "gruvbox"

// Theme contains semantic color tokens shared by every TUI component.
// Values use Lip Gloss color syntax and are intentionally kept private to the
// built-in registry so components never depend on palette-specific names.
type Theme struct {
	Name        string
	Canvas      string
	Panel       string
	Selection   string
	Text        string
	Muted       string
	Placeholder string
	Border      string
	Accent      string
	Success     string
	Warning     string
	Error       string
	Info        string
}

var themeOrder = []string{"gruvbox", "tokyo-night"}

var builtInThemes = map[string]Theme{
	"gruvbox": {
		Name:        "gruvbox",
		Canvas:      "#282828",
		Panel:       "#3c3836",
		Selection:   "#504945",
		Text:        "#ebdbb2",
		Muted:       "#bdae93",
		Placeholder: "#a89984",
		Border:      "#665c54",
		Accent:      "#fe8019",
		Success:     "#b8bb26",
		Warning:     "#fabd2f",
		Error:       "#fb4934",
		Info:        "#83a598",
	},
	"tokyo-night": {
		Name:        "tokyo-night",
		Canvas:      "#1a1b26",
		Panel:       "#24283b",
		Selection:   "#33467c",
		Text:        "#c0caf5",
		Muted:       "#a9b1d6",
		Placeholder: "#565f89",
		Border:      "#414868",
		Accent:      "#7aa2f7",
		Success:     "#9ece6a",
		Warning:     "#e0af68",
		Error:       "#f7768e",
		Info:        "#7dcfff",
	},
}

// AvailableThemes returns valid theme names in stable display order.
func AvailableThemes() []string {
	return append([]string(nil), themeOrder...)
}

// ResolveTheme resolves a built-in theme. An empty name selects Gruvbox.
func ResolveTheme(name string) (Theme, error) {
	if name == "" {
		name = DefaultThemeName
	}
	theme, ok := builtInThemes[name]
	if !ok {
		return Theme{}, fmt.Errorf("unknown theme %q (available: gruvbox, tokyo-night)", name)
	}
	return theme, nil
}
