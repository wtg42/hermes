package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type commandMode uint8

const (
	commandModeNormal commandMode = iota
	commandModeRoot
	commandModeTemplate
	commandModeConfirmClear
	commandModeConfirmQuit
	commandModeHelp
)

func (m commandMode) String() string {
	switch m {
	case commandModeNormal:
		return "normal"
	case commandModeRoot:
		return "root"
	case commandModeTemplate:
		return "template"
	case commandModeConfirmClear:
		return "confirm-clear"
	case commandModeConfirmQuit:
		return "confirm-quit"
	case commandModeHelp:
		return "help"
	default:
		return "unknown"
	}
}

type commandID string

const (
	commandNone         commandID = ""
	commandAttach       commandID = "attach"
	commandTemplates    commandID = "templates"
	commandClear        commandID = "clear"
	commandConfirmClear commandID = "confirm-clear"
	commandQuit         commandID = "quit"
	commandConfirmQuit  commandID = "confirm-quit"
	commandHelp         commandID = "help"
	commandTemplateHTML commandID = "template-html"
	commandTemplateText commandID = "template-text"
	commandTemplateEML  commandID = "template-eml"
)

type commandDefinition struct {
	ID    commandID
	Key   string
	Label string
}

var commandRegistry = map[commandMode][]commandDefinition{
	commandModeRoot: {
		{ID: commandAttach, Key: "a", Label: "Attach"},
		{ID: commandTemplates, Key: "t", Label: "Template"},
		{ID: commandClear, Key: "c", Label: "Clear"},
		{ID: commandQuit, Key: "q", Label: "Quit"},
		{ID: commandHelp, Key: "?", Label: "Help"},
	},
	commandModeTemplate: {
		{ID: commandTemplateHTML, Key: "h", Label: "HTML"},
		{ID: commandTemplateText, Key: "p", Label: "Plain Text"},
		{ID: commandTemplateEML, Key: "e", Label: "EML"},
	},
}

type commandPrefix struct {
	mode   commandMode
	notice string
}

func newCommandPrefix() commandPrefix {
	return commandPrefix{mode: commandModeNormal}
}

func (p commandPrefix) Update(msg tea.Msg) (commandPrefix, commandID, bool) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, commandNone, false
	}

	key := keyMsg.String()
	if p.mode == commandModeNormal {
		p.notice = ""
		if key == "ctrl+x" {
			p.mode = commandModeRoot
			return p, commandNone, true
		}
		return p, commandNone, false
	}

	switch p.mode {
	case commandModeRoot:
		if key == "esc" {
			p.mode = commandModeNormal
			return p, commandNone, true
		}
		definition, found := lookupCommand(commandModeRoot, key)
		if !found {
			return p.unknown(key)
		}
		switch definition.ID {
		case commandTemplates:
			p.mode = commandModeTemplate
		case commandHelp:
			p.mode = commandModeHelp
		default:
			p.mode = commandModeNormal
		}
		return p, definition.ID, true

	case commandModeTemplate:
		if key == "esc" {
			p.mode = commandModeRoot
			return p, commandNone, true
		}
		definition, found := lookupCommand(commandModeTemplate, key)
		if !found {
			return p.unknown(key)
		}
		p.mode = commandModeNormal
		return p, definition.ID, true

	case commandModeConfirmClear:
		if key == "esc" {
			p.mode = commandModeNormal
			return p, commandNone, true
		}
		if key == "c" {
			p.mode = commandModeNormal
			return p, commandConfirmClear, true
		}
		return p.unknown(key)

	case commandModeConfirmQuit:
		if key == "esc" {
			p.mode = commandModeNormal
			return p, commandNone, true
		}
		if key == "q" {
			p.mode = commandModeNormal
			return p, commandConfirmQuit, true
		}
		return p.unknown(key)

	case commandModeHelp:
		if key == "esc" {
			p.mode = commandModeNormal
		}
		return p, commandNone, true

	default:
		p.mode = commandModeNormal
		return p, commandNone, false
	}
}

func (p commandPrefix) unknown(key string) (commandPrefix, commandID, bool) {
	p.mode = commandModeNormal
	p.notice = fmt.Sprintf("Unknown command: %s", key)
	return p, commandNone, true
}

func lookupCommand(mode commandMode, key string) (commandDefinition, bool) {
	for _, definition := range commandRegistry[mode] {
		if definition.Key == key {
			return definition, true
		}
	}
	return commandDefinition{}, false
}

func renderCommandHUD(mode commandMode) string {
	var title string
	var cancelLabel string
	switch mode {
	case commandModeRoot:
		title = "COMMAND"
		cancelLabel = "[Esc] Cancel"
	case commandModeTemplate:
		title = "TEMPLATE"
		cancelLabel = "[Esc] Back"
	default:
		return ""
	}

	parts := []string{title}
	for _, definition := range commandRegistry[mode] {
		parts = append(parts, fmt.Sprintf("[%s] %s", strings.ToUpper(definition.Key), definition.Label))
	}
	parts = append(parts, cancelLabel)
	return strings.Join(parts, "  ")
}

func renderCommandHelp() string {
	return strings.Join([]string{
		"Commands",
		"",
		"Direct",
		"  Ctrl+S      Send",
		"  Ctrl+J/K    Switch panel",
		"  Tab         Next field",
		"  Shift+Tab   Previous field",
		"",
		"Prefix",
		"  Ctrl+X      Open commands",
		renderCommandHUD(commandModeRoot),
		"",
		"[Esc] Close",
	}, "\n")
}
