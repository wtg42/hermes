package tui

import (
	"reflect"
	"strings"
	"testing"
)

func TestAvailableThemesAreStableAndLimited(t *testing.T) {
	want := []string{"gruvbox", "tokyo-night"}
	if got := AvailableThemes(); !reflect.DeepEqual(got, want) {
		t.Fatalf("AvailableThemes() = %v, want %v", got, want)
	}

	// Callers must not be able to mutate registry order through the returned slice.
	got := AvailableThemes()
	got[0] = "changed"
	if next := AvailableThemes(); !reflect.DeepEqual(next, want) {
		t.Fatalf("AvailableThemes() exposed mutable registry state: %v", next)
	}
}

func TestResolveTheme(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTheme string
	}{
		{name: "empty uses default", wantTheme: "gruvbox"},
		{name: "gruvbox", input: "gruvbox", wantTheme: "gruvbox"},
		{name: "tokyo night", input: "tokyo-night", wantTheme: "tokyo-night"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme, err := ResolveTheme(tt.input)
			if err != nil {
				t.Fatalf("ResolveTheme(%q) returned error: %v", tt.input, err)
			}
			if theme.Name != tt.wantTheme {
				t.Fatalf("ResolveTheme(%q).Name = %q, want %q", tt.input, theme.Name, tt.wantTheme)
			}
		})
	}
}

func TestResolveThemeRejectsUnknownName(t *testing.T) {
	_, err := ResolveTheme("nord")
	if err == nil {
		t.Fatal("ResolveTheme(nord) unexpectedly succeeded")
	}
	for _, want := range []string{"nord", "gruvbox", "tokyo-night"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not mention %q", err, want)
		}
	}
}

func TestBuiltInThemesProvideEverySemanticToken(t *testing.T) {
	for _, name := range AvailableThemes() {
		t.Run(name, func(t *testing.T) {
			theme, err := ResolveTheme(name)
			if err != nil {
				t.Fatal(err)
			}
			value := reflect.ValueOf(theme)
			typeOfTheme := value.Type()
			for i := 0; i < value.NumField(); i++ {
				field := typeOfTheme.Field(i)
				if field.Name == "Name" {
					continue
				}
				if got := value.Field(i).String(); got == "" {
					t.Errorf("theme %q token %s is empty", name, field.Name)
				}
			}
		})
	}
}

func TestBuiltInThemeRepresentativeColors(t *testing.T) {
	gruvbox, _ := ResolveTheme("gruvbox")
	if gruvbox.Canvas != "#282828" || gruvbox.Panel != "#3c3836" || gruvbox.Border != "#504945" || gruvbox.Accent != "#d79921" {
		t.Fatalf("unexpected Gruvbox palette: %+v", gruvbox)
	}

	tokyoNight, _ := ResolveTheme("tokyo-night")
	if tokyoNight.Canvas != "#1a1b26" || tokyoNight.Panel != "#24283b" || tokyoNight.Border != "#3b4261" || tokyoNight.Accent != "#7aa2f7" {
		t.Fatalf("unexpected Tokyo Night palette: %+v", tokyoNight)
	}
}
