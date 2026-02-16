package ssh

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Client manages SSH connections to a Soft Serve instance.
type Client struct {
	host      string
	port      int
	username  string
	signer    ssh.Signer
	agentConn net.Conn
	agentAuth ssh.AuthMethod
}

// ClientConfig holds configuration for creating a new SSH client.
type ClientConfig struct {
	Host           string
	Port           int
	Username       string
	PrivateKey     string // PEM-encoded private key contents
	PrivateKeyPath string // Path to private key file
	UseAgent       bool
	IdentityFile   string // Path to public key file to filter agent keys
}

// NewClient creates a new SSH client for Soft Serve.
func NewClient(cfg ClientConfig) (*Client, error) {
	c := &Client{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
	}

	// Try private key first (takes precedence)
	if cfg.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("parsing private key: %w", err)
		}
		c.signer = signer
	} else if cfg.PrivateKeyPath != "" {
		keyData, err := os.ReadFile(cfg.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("reading private key file %s: %w", cfg.PrivateKeyPath, err)
		}
		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("parsing private key from %s: %w", cfg.PrivateKeyPath, err)
		}
		c.signer = signer
	}

	// Set up SSH agent if requested
	if cfg.UseAgent {
		socket := os.Getenv("SSH_AUTH_SOCK")
		if socket != "" {
			conn, err := net.Dial("unix", socket)
			if err == nil {
				c.agentConn = conn
				agentClient := agent.NewClient(conn)
				if cfg.IdentityFile != "" {
					c.agentAuth, err = filteredAgentAuth(agentClient, cfg.IdentityFile)
					if err != nil {
						_ = conn.Close()
						return nil, fmt.Errorf("filtering agent keys with identity file: %w", err)
					}
				} else {
					c.agentAuth = ssh.PublicKeysCallback(agentClient.Signers)
				}
			}
		}
	}

	if c.signer == nil && c.agentAuth == nil {
		return nil, fmt.Errorf("no authentication method available: provide a private key or enable SSH agent")
	}

	return c, nil
}

// Close cleans up any resources held by the client.
func (c *Client) Close() error {
	if c.agentConn != nil {
		return c.agentConn.Close()
	}
	return nil
}

// filteredAgentAuth reads a public key from identityFile and returns an
// AuthMethod that only offers the matching key from the SSH agent. This
// mirrors OpenSSH's IdentityFile behavior when used with an agent.
func filteredAgentAuth(agentClient agent.ExtendedAgent, identityFile string) (ssh.AuthMethod, error) {
	pubKeyData, err := os.ReadFile(identityFile)
	if err != nil {
		return nil, fmt.Errorf("reading identity file %s: %w", identityFile, err)
	}
	wantKey, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyData)
	if err != nil {
		return nil, fmt.Errorf("parsing public key from %s: %w", identityFile, err)
	}
	wantBytes := wantKey.Marshal()

	return ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
		signers, err := agentClient.Signers()
		if err != nil {
			return nil, err
		}
		for _, s := range signers {
			if bytes.Equal(s.PublicKey().Marshal(), wantBytes) {
				return []ssh.Signer{s}, nil
			}
		}
		return nil, fmt.Errorf("identity file %s: matching key not found in SSH agent", identityFile)
	}), nil
}

// Run executes a command on the Soft Serve server and returns stdout.
func (c *Client) Run(command string) (string, error) {
	var authMethods []ssh.AuthMethod
	if c.signer != nil {
		authMethods = append(authMethods, ssh.PublicKeys(c.signer))
	}
	if c.agentAuth != nil {
		authMethods = append(authMethods, c.agentAuth)
	}

	config := &ssh.ClientConfig{
		User:            c.username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // Soft Serve doesn't typically use host key verification
	}

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", fmt.Errorf("connecting to %s: %w", addr, err)
	}
	defer func() { _ = conn.Close() }()

	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("creating session: %w", err)
	}
	defer func() { _ = session.Close() }()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(command); err != nil {
		return "", fmt.Errorf("running command %q: %s: %w", command, strings.TrimSpace(stderr.String()), err)
	}

	return strings.TrimRight(stdout.String(), "\n"), nil
}

// RepoCreate creates a new repository.
func (c *Client) RepoCreate(name string, opts RepoCreateOpts) error {
	cmd := fmt.Sprintf("repo create %s", name)
	if opts.Description != "" {
		cmd += fmt.Sprintf(" -d %q", opts.Description)
	}
	if opts.ProjectName != "" {
		cmd += fmt.Sprintf(" -n %q", opts.ProjectName)
	}
	if opts.Private {
		cmd += " -p"
	}
	_, err := c.Run(cmd)
	return err
}

// RepoCreateOpts holds options for creating a repository.
type RepoCreateOpts struct {
	Description string
	ProjectName string
	Private     bool
}

// RepoInfo retrieves information about a repository.
func (c *Client) RepoInfo(name string) (*RepoInfoResult, error) {
	output, err := c.Run(fmt.Sprintf("repo info %s", name))
	if err != nil {
		return nil, err
	}
	return ParseRepoInfo(output)
}

// RepoDelete deletes a repository.
func (c *Client) RepoDelete(name string) error {
	_, err := c.Run(fmt.Sprintf("repo delete %s", name))
	return err
}

// RepoSetDescription sets a repository's description.
func (c *Client) RepoSetDescription(name, description string) error {
	_, err := c.Run(fmt.Sprintf("repo description %s %q", name, description))
	return err
}

// RepoSetPrivate sets whether a repository is private.
func (c *Client) RepoSetPrivate(name string, private bool) error {
	_, err := c.Run(fmt.Sprintf("repo private %s %t", name, private))
	return err
}

// RepoSetHidden sets whether a repository is hidden.
func (c *Client) RepoSetHidden(name string, hidden bool) error {
	_, err := c.Run(fmt.Sprintf("repo hidden %s %t", name, hidden))
	return err
}

// RepoSetProjectName sets a repository's project name.
func (c *Client) RepoSetProjectName(name, projectName string) error {
	_, err := c.Run(fmt.Sprintf("repo project-name %s %q", name, projectName))
	return err
}

// UserCreate creates a new user.
func (c *Client) UserCreate(username string, opts UserCreateOpts) error {
	cmd := fmt.Sprintf("user create %s", username)
	if opts.Admin {
		cmd += " -a"
	}
	for _, key := range opts.PublicKeys {
		cmd += fmt.Sprintf(" -k %q", key)
	}
	_, err := c.Run(cmd)
	return err
}

// UserCreateOpts holds options for creating a user.
type UserCreateOpts struct {
	Admin      bool
	PublicKeys []string
}

// UserInfo retrieves information about a user.
func (c *Client) UserInfo(username string) (*UserInfoResult, error) {
	output, err := c.Run(fmt.Sprintf("user info %s", username))
	if err != nil {
		return nil, err
	}
	return ParseUserInfo(output)
}

// UserDelete deletes a user.
func (c *Client) UserDelete(username string) error {
	_, err := c.Run(fmt.Sprintf("user delete %s", username))
	return err
}

// UserSetAdmin sets whether a user is an admin.
func (c *Client) UserSetAdmin(username string, admin bool) error {
	_, err := c.Run(fmt.Sprintf("user set-admin %s %t", username, admin))
	return err
}

// UserAddPublicKey adds a public key to a user.
func (c *Client) UserAddPublicKey(username, key string) error {
	_, err := c.Run(fmt.Sprintf("user add-pubkey %s %q", username, key))
	return err
}

// UserRemovePublicKey removes a public key from a user.
func (c *Client) UserRemovePublicKey(username, key string) error {
	_, err := c.Run(fmt.Sprintf("user remove-pubkey %s %q", username, key))
	return err
}

// CollabAdd adds a collaborator to a repository.
func (c *Client) CollabAdd(repo, username, accessLevel string) error {
	cmd := fmt.Sprintf("repo collab add %s %s", repo, username)
	if accessLevel != "" {
		cmd += " " + accessLevel
	}
	_, err := c.Run(cmd)
	return err
}

// CollabList lists collaborators for a repository.
func (c *Client) CollabList(repo string) ([]CollabEntry, error) {
	output, err := c.Run(fmt.Sprintf("repo collab list %s", repo))
	if err != nil {
		return nil, err
	}
	return ParseCollabList(output)
}

// CollabRemove removes a collaborator from a repository.
func (c *Client) CollabRemove(repo, username string) error {
	_, err := c.Run(fmt.Sprintf("repo collab remove %s %s", repo, username))
	return err
}

// SettingsGetAllowKeyless gets the allow-keyless setting.
func (c *Client) SettingsGetAllowKeyless() (bool, error) {
	output, err := c.Run("settings allow-keyless")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) == "true", nil
}

// SettingsSetAllowKeyless sets the allow-keyless setting.
func (c *Client) SettingsSetAllowKeyless(allow bool) error {
	_, err := c.Run(fmt.Sprintf("settings allow-keyless %t", allow))
	return err
}

// SettingsGetAnonAccess gets the anonymous access level.
func (c *Client) SettingsGetAnonAccess() (string, error) {
	output, err := c.Run("settings anon-access")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// SettingsSetAnonAccess sets the anonymous access level.
func (c *Client) SettingsSetAnonAccess(level string) error {
	_, err := c.Run(fmt.Sprintf("settings anon-access %s", level))
	return err
}
