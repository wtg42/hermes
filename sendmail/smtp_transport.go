package sendmail

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// SMTP transport modes are closed sets accepted by the structured CLI.
const (
	TLSModeNone     = "none"
	TLSModeRequired = "required"
	AuthModeNone    = "none"
	AuthModePlain   = "plain"
)

// SMTPTransportConfig contains connection policy and one-shot credentials.
// Password and RootCAs are runtime-only and must never be persisted.
type SMTPTransportConfig struct {
	Server         string
	Port           string
	TLSMode        string
	TLSServerName  string
	AuthMode       string
	AuthUsername   string
	PasswordSource bool
	Password       string
	RootCAs        *x509.CertPool
}

// DefaultSMTPTransportConfig returns the backward-compatible plaintext policy.
func DefaultSMTPTransportConfig(server, port string) SMTPTransportConfig {
	return SMTPTransportConfig{Server: server, Port: port, TLSMode: TLSModeNone, AuthMode: AuthModeNone}
}

// ValidateStatic validates transport policy before reading a password or dialing.
func (config SMTPTransportConfig) ValidateStatic() error {
	if config.TLSMode != TLSModeNone && config.TLSMode != TLSModeRequired {
		return fmt.Errorf("tls mode must be %q or %q", TLSModeNone, TLSModeRequired)
	}
	if config.AuthMode != AuthModeNone && config.AuthMode != AuthModePlain {
		return fmt.Errorf("auth mode must be %q or %q", AuthModeNone, AuthModePlain)
	}
	if config.TLSMode == TLSModeRequired {
		if strings.TrimSpace(config.TLSServerName) == "" || strings.ContainsAny(config.TLSServerName, "\r\n\x00") {
			return fmt.Errorf("tls server name is required for required TLS")
		}
	} else if config.TLSServerName != "" {
		return fmt.Errorf("tls server name requires tls mode %q", TLSModeRequired)
	}
	if config.AuthMode == AuthModePlain {
		if config.TLSMode != TLSModeRequired {
			return fmt.Errorf("plain auth requires tls mode %q", TLSModeRequired)
		}
		if strings.TrimSpace(config.AuthUsername) == "" {
			return fmt.Errorf("auth username is required for plain auth")
		}
		if !config.PasswordSource {
			return fmt.Errorf("plain auth requires --auth-password-stdin")
		}
	} else {
		if config.AuthUsername != "" {
			return fmt.Errorf("auth username requires auth mode %q", AuthModePlain)
		}
		if config.PasswordSource {
			return fmt.Errorf("--auth-password-stdin requires auth mode %q", AuthModePlain)
		}
	}
	return nil
}

func (config SMTPTransportConfig) validateReady() error {
	if err := config.ValidateStatic(); err != nil {
		return err
	}
	if config.AuthMode == AuthModePlain && config.Password == "" {
		return fmt.Errorf("SMTP password from stdin must not be empty")
	}
	return nil
}

func sendSMTPWithTransport(config SMTPTransportConfig, from string, recipients []string, message []byte) error {
	if err := config.validateReady(); err != nil {
		return fmt.Errorf("transport preflight: %w", err)
	}
	if config.TLSMode == TLSModeNone {
		if err := SendMail(net.JoinHostPort(config.Server, config.Port), nil, from, recipients, message); err != nil {
			return fmt.Errorf("SMTP delivery stage: %w", err)
		}
		return nil
	}

	address := net.JoinHostPort(config.Server, config.Port)
	connection, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("TCP dial stage: %w", err)
	}
	client, err := smtp.NewClient(connection, config.TLSServerName)
	if err != nil {
		_ = connection.Close()
		return fmt.Errorf("SMTP greeting stage: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); !ok {
		return fmt.Errorf("STARTTLS stage: server does not advertise STARTTLS")
	}
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: config.TLSServerName,
		RootCAs:    config.RootCAs,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("TLS verification stage: %w", err)
	}
	if config.AuthMode == AuthModePlain {
		auth := smtp.PlainAuth("", config.AuthUsername, config.Password, config.TLSServerName)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth stage: authentication rejected")
		}
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP envelope MAIL stage: %w", err)
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("SMTP envelope RCPT stage: %w", err)
		}
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA stage: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("SMTP DATA write stage: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("SMTP DATA close stage: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("SMTP quit stage: %w", err)
	}
	return nil
}
