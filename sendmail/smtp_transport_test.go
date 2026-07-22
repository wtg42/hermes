package sendmail

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"
)

func TestSMTPTransportStaticPreflight(t *testing.T) {
	tests := []struct {
		name string
		cfg  SMTPTransportConfig
		ok   bool
	}{
		{"plaintext default", DefaultSMTPTransportConfig("192.0.2.10", "25"), true},
		{"required TLS", SMTPTransportConfig{TLSMode: TLSModeRequired, TLSServerName: "smtp.test", AuthMode: AuthModeNone}, true},
		{"invalid TLS", SMTPTransportConfig{TLSMode: "auto", AuthMode: AuthModeNone}, false},
		{"required without name", SMTPTransportConfig{TLSMode: TLSModeRequired, AuthMode: AuthModeNone}, false},
		{"unused name", SMTPTransportConfig{TLSMode: TLSModeNone, TLSServerName: "smtp.test", AuthMode: AuthModeNone}, false},
		{"plain over plaintext", SMTPTransportConfig{TLSMode: TLSModeNone, AuthMode: AuthModePlain, AuthUsername: "user", PasswordSource: true}, false},
		{"plain missing source", SMTPTransportConfig{TLSMode: TLSModeRequired, TLSServerName: "smtp.test", AuthMode: AuthModePlain, AuthUsername: "user"}, false},
		{"plain ready source", SMTPTransportConfig{TLSMode: TLSModeRequired, TLSServerName: "smtp.test", AuthMode: AuthModePlain, AuthUsername: "user", PasswordSource: true}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.cfg.ValidateStatic()
			if (err == nil) != test.ok {
				t.Fatalf("ValidateStatic() error = %v, ok=%v", err, test.ok)
			}
		})
	}
}

func TestSMTPRequiredTLSAndPlainAuthSequence(t *testing.T) {
	address, roots, events, stop := startSMTPTransportTestServer(t, true, true)
	defer stop()
	host, port, _ := net.SplitHostPort(address)
	config := SMTPTransportConfig{Server: host, Port: port, TLSMode: TLSModeRequired, TLSServerName: "smtp.test", AuthMode: AuthModePlain, AuthUsername: "user", PasswordSource: true, Password: "secret-value", RootCAs: roots}
	if err := sendSMTPWithTransport(config, "from@example.com", []string{"to@example.com"}, []byte("Subject: test\r\n\r\nbody")); err != nil {
		t.Fatalf("sendSMTPWithTransport() error = %v", err)
	}
	joined := strings.Join(*events, " ")
	for _, expected := range []string{"STARTTLS", "AUTH", "MAIL", "RCPT", "DATA"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("events %q missing %s", joined, expected)
		}
	}
	if strings.Index(joined, "AUTH") > strings.Index(joined, "MAIL") {
		t.Fatalf("auth occurred after envelope: %q", joined)
	}
}

func TestSMTPRequiredTLSDoesNotDowngrade(t *testing.T) {
	address, _, events, stop := startSMTPTransportTestServer(t, false, false)
	defer stop()
	host, port, _ := net.SplitHostPort(address)
	err := sendSMTPWithTransport(SMTPTransportConfig{Server: host, Port: port, TLSMode: TLSModeRequired, TLSServerName: "smtp.test", AuthMode: AuthModeNone}, "from@example.com", []string{"to@example.com"}, []byte("body"))
	if err == nil || !strings.Contains(err.Error(), "STARTTLS") || strings.Contains(strings.Join(*events, " "), "MAIL") {
		t.Fatalf("error=%v events=%v", err, *events)
	}
}

func startSMTPTransportTestServer(t *testing.T, advertiseTLS, acceptAuth bool) (string, *x509.CertPool, *[]string, func()) {
	t.Helper()
	certificate, roots := testTLSCertificate(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("sandbox does not allow loopback listeners: %v", err)
		}
		t.Fatal(err)
	}
	events := &[]string{}
	done := make(chan struct{})
	go func() {
		defer close(done)
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer connection.Close()
		reader := bufio.NewReader(connection)
		writer := bufio.NewWriter(connection)
		write := func(value string) { _, _ = writer.WriteString(value + "\r\n"); _ = writer.Flush() }
		write("220 smtp.test ESMTP")
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				return
			}
			command := strings.TrimSpace(line)
			upper := strings.ToUpper(command)
			switch {
			case strings.HasPrefix(upper, "EHLO"):
				*events = append(*events, "EHLO")
				if advertiseTLS {
					write("250-smtp.test")
					write("250-STARTTLS")
					write("250 AUTH PLAIN")
				} else {
					write("250 smtp.test")
				}
			case upper == "STARTTLS":
				*events = append(*events, "STARTTLS")
				write("220 ready")
				tlsConnection := tls.Server(connection, &tls.Config{Certificates: []tls.Certificate{certificate}, MinVersion: tls.VersionTLS12})
				if tlsConnection.Handshake() != nil {
					return
				}
				connection = tlsConnection
				reader = bufio.NewReader(connection)
				writer = bufio.NewWriter(connection)
			case strings.HasPrefix(upper, "AUTH PLAIN"):
				*events = append(*events, "AUTH")
				if acceptAuth {
					write("235 authenticated")
				} else {
					write("535 rejected")
				}
			case strings.HasPrefix(upper, "MAIL FROM"):
				*events = append(*events, "MAIL")
				write("250 ok")
			case strings.HasPrefix(upper, "RCPT TO"):
				*events = append(*events, "RCPT")
				write("250 ok")
			case upper == "DATA":
				*events = append(*events, "DATA")
				write("354 send")
				for {
					value, _ := reader.ReadString('\n')
					if strings.TrimSpace(value) == "." {
						break
					}
				}
				write("250 queued")
			case upper == "QUIT":
				write("221 bye")
				return
			default:
				write("250 ok")
			}
		}
	}()
	return listener.Addr().String(), roots, events, func() { _ = listener.Close(); <-done }
}

func testTLSCertificate(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	template := x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "smtp.test"}, DNSNames: []string{"smtp.test"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, IsCA: true, BasicConstraintsValid: true}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certificate, err := tls.X509KeyPair(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(parsed)
	return certificate, roots
}
