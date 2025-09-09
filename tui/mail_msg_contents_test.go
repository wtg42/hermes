package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func TestMailMsgModel_HotkeyInsertHTMLTemplate(t *testing.T) {
	// Initialize model
	m := MailMsgModel{
		textarea:        textarea.New(),
		whichOneOnFocus: 1, // Focus on textarea
	}

	// Create a KeyMsg for ctrl+h
	msg := tea.KeyMsg{
		Type: tea.KeyCtrlH,
	}

	updatedModel, _ := m.Update(msg)

	// Type assert to MailMsgModel
	mm, ok := updatedModel.(MailMsgModel)
	if !ok {
		t.Fatalf("Expected MailMsgModel, got %T", updatedModel)
	}

	// Check if HTML template was inserted
	expectedHTML := `<html>
<head>
    <title>Email Template</title>
</head>
<body>
    <h1>Hello!</h1>
    <p>This is an HTML email template.</p>
    <p>Best regards,<br>Your Name</p>
</body>
</html>`

	if mm.textarea.Value() != expectedHTML {
		t.Errorf("Expected HTML template to be inserted, got: %s", mm.textarea.Value())
	}
}

func TestMailMsgModel_HotkeyInsertTextTemplate(t *testing.T) {
	// Initialize model
	m := MailMsgModel{
		textarea:        textarea.New(),
		whichOneOnFocus: 1, // Focus on textarea
	}

	// Create a KeyMsg for ctrl+t
	msg := tea.KeyMsg{
		Type: tea.KeyCtrlT,
	}

	updatedModel, _ := m.Update(msg)

	// Type assert to MailMsgModel
	mm, ok := updatedModel.(MailMsgModel)
	if !ok {
		t.Fatalf("Expected MailMsgModel, got %T", updatedModel)
	}

	// Check if text template was inserted
	expectedText := `Hello,

This is a plain text email template.

Best regards,
Your Name`

	if mm.textarea.Value() != expectedText {
		t.Errorf("Expected text template to be inserted, got: %s", mm.textarea.Value())
	}
}

func TestMailMsgModel_HotkeyInsertEMLTemplate(t *testing.T) {
	// Initialize model
	m := MailMsgModel{
		textarea:        textarea.New(),
		whichOneOnFocus: 1, // Focus on textarea
	}

	// Create a KeyMsg for ctrl+e
	msg := tea.KeyMsg{
		Type: tea.KeyCtrlE,
	}

	updatedModel, _ := m.Update(msg)

	// Type assert to MailMsgModel
	mm, ok := updatedModel.(MailMsgModel)
	if !ok {
		t.Fatalf("Expected MailMsgModel, got %T", updatedModel)
	}

	// Check if EML template was inserted
	expectedEML := `Return-Path: <sender@example.com>
Received: by smtp.example.com id 123456; Mon, 1 Jan 2024 12:00:00 +0000
Date: Mon, 1 Jan 2024 12:00:00 +0000
From: Sender Name <sender@example.com>
To: Recipient Name <recipient@example.com>
Subject: Test Email
Content-Type: text/plain; charset=UTF-8

Hello,

This is a sample EML email content.

Best regards,
Sender Name`

	if mm.textarea.Value() != expectedEML {
		t.Errorf("Expected EML template to be inserted, got: %s", mm.textarea.Value())
	}
}

func TestMailMsgModel_HotkeysOnlyWorkInTextareaFocus(t *testing.T) {
	// Initialize model with filepicker focus
	m := MailMsgModel{
		textarea:        textarea.New(),
		whichOneOnFocus: 2, // Focus on filepicker
	}

	// Set some initial content
	m.textarea.SetValue("Initial content")

	// Create a KeyMsg for ctrl+h
	msg := tea.KeyMsg{
		Type: tea.KeyCtrlH,
	}

	updatedModel, _ := m.Update(msg)

	// Type assert to MailMsgModel
	mm, ok := updatedModel.(MailMsgModel)
	if !ok {
		t.Fatalf("Expected MailMsgModel, got %T", updatedModel)
	}

	// Content should remain unchanged
	if mm.textarea.Value() != "Initial content" {
		t.Errorf("Expected content to remain unchanged when not in textarea focus, got: %s", mm.textarea.Value())
	}
}
