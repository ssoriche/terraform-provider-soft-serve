package ssh

import (
	"os"
	"testing"
)

func TestNewClient_NoAuthMethod(t *testing.T) {
	// Ensure SSH agent is unavailable
	t.Setenv("SSH_AUTH_SOCK", "")

	_, err := NewClient(ClientConfig{
		Host:     "localhost",
		Port:     23231,
		Username: "admin",
		UseAgent: false,
	})

	if err == nil {
		t.Fatal("expected error when no authentication method is available")
	}
}

func TestNewClient_InvalidPrivateKey(t *testing.T) {
	_, err := NewClient(ClientConfig{
		Host:       "localhost",
		Port:       23231,
		Username:   "admin",
		PrivateKey: "not-a-valid-key",
	})

	if err == nil {
		t.Fatal("expected error for invalid private key")
	}
}

func TestNewClient_PrivateKeyFileNotFound(t *testing.T) {
	_, err := NewClient(ClientConfig{
		Host:           "localhost",
		Port:           23231,
		Username:       "admin",
		PrivateKeyPath: "/nonexistent/path/to/key",
	})

	if err == nil {
		t.Fatal("expected error for nonexistent private key file")
	}
}

func TestNewClient_InvalidPrivateKeyFile(t *testing.T) {
	// Create a temp file with invalid key content
	tmpFile, err := os.CreateTemp("", "test-key-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

	if _, err := tmpFile.WriteString("not-a-valid-key"); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = NewClient(ClientConfig{
		Host:           "localhost",
		Port:           23231,
		Username:       "admin",
		PrivateKeyPath: tmpFile.Name(),
	})

	if err == nil {
		t.Fatal("expected error for invalid private key file content")
	}
}

func TestNewClient_InvalidIdentityFile(t *testing.T) {
	// Ensure SSH agent socket is available for this test
	origSock := os.Getenv("SSH_AUTH_SOCK")
	if origSock == "" {
		t.Skip("SSH_AUTH_SOCK not set, skipping identity file test")
	}

	_, err := NewClient(ClientConfig{
		Host:         "localhost",
		Port:         23231,
		Username:     "admin",
		UseAgent:     true,
		IdentityFile: "/nonexistent/identity/file",
	})

	if err == nil {
		t.Fatal("expected error for nonexistent identity file")
	}
}

func TestNewClient_InvalidIdentityFileContent(t *testing.T) {
	origSock := os.Getenv("SSH_AUTH_SOCK")
	if origSock == "" {
		t.Skip("SSH_AUTH_SOCK not set, skipping identity file content test")
	}

	tmpFile, err := os.CreateTemp("", "test-identity-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

	if _, err := tmpFile.WriteString("not-a-valid-public-key"); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = NewClient(ClientConfig{
		Host:         "localhost",
		Port:         23231,
		Username:     "admin",
		UseAgent:     true,
		IdentityFile: tmpFile.Name(),
	})

	if err == nil {
		t.Fatal("expected error for invalid identity file content")
	}
}

func TestClientClose_NilAgentConn(t *testing.T) {
	c := &Client{
		host:     "localhost",
		port:     23231,
		username: "admin",
	}

	err := c.Close()
	if err != nil {
		t.Errorf("Close() with nil agent conn should not error, got: %v", err)
	}
}
