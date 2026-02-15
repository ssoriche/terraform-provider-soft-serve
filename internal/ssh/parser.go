package ssh

import (
	"fmt"
	"strings"
)

// RepoInfoResult holds parsed repository information.
type RepoInfoResult struct {
	ProjectName string
	Repository  string
	Description string
	Private     bool
	Hidden      bool
	Mirror      bool
	Owner       string
}

// UserInfoResult holds parsed user information.
type UserInfoResult struct {
	Username   string
	Admin      bool
	PublicKeys []string
}

// CollabEntry holds a parsed collaborator entry.
type CollabEntry struct {
	Username    string
	AccessLevel string
}

// ParseRepoInfo parses the output of `repo info <name>`.
//
// Expected format:
//
//	Project Name: myproject
//	Repository: myrepo
//	Description: A test repo
//	Private: false
//	Hidden: false
//	Mirror: false
//	Owner: admin
//	Default Branch: main
//	Branches:
//	  - main
//	Tags:
func ParseRepoInfo(output string) (*RepoInfoResult, error) {
	result := &RepoInfoResult{}
	kvs := parseKeyValues(output)

	for _, kv := range kvs {
		switch kv.key {
		case "Project Name":
			result.ProjectName = kv.value
		case "Repository":
			result.Repository = kv.value
		case "Description":
			result.Description = kv.value
		case "Private":
			result.Private = kv.value == "true"
		case "Hidden":
			result.Hidden = kv.value == "true"
		case "Mirror":
			result.Mirror = kv.value == "true"
		case "Owner":
			result.Owner = kv.value
		}
	}

	if result.Repository == "" {
		return nil, fmt.Errorf("failed to parse repo info: missing Repository field")
	}

	return result, nil
}

// ParseUserInfo parses the output of `user info <username>`.
//
// Expected format:
//
//	Username: alice
//	Admin: false
//	Public keys:
//	  ssh-ed25519 AAAA... alice@host
//	  ssh-rsa AAAA... alice@other
func ParseUserInfo(output string) (*UserInfoResult, error) {
	result := &UserInfoResult{}
	lines := strings.Split(output, "\n")

	inPublicKeys := false
	for _, line := range lines {
		if inPublicKeys {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				// Check if this is a new key-value line (not indented content)
				if !strings.HasPrefix(line, "  ") && strings.Contains(line, ": ") {
					inPublicKeys = false
					// Fall through to key-value parsing below
				} else {
					result.PublicKeys = append(result.PublicKeys, trimmed)
					continue
				}
			} else {
				continue
			}
		}

		key, value, ok := parseKeyValue(line)
		if !ok {
			continue
		}

		switch key {
		case "Username":
			result.Username = value
		case "Admin":
			result.Admin = value == "true"
		case "Public keys":
			inPublicKeys = true
		}
	}

	if result.Username == "" {
		return nil, fmt.Errorf("failed to parse user info: missing Username field")
	}

	return result, nil
}

// ParseCollabList parses the output of `repo collab list <repo>`.
//
// Expected format (one entry per line):
//
//	alice read-write
//	bob read-only
func ParseCollabList(output string) ([]CollabEntry, error) {
	if strings.TrimSpace(output) == "" {
		return nil, nil
	}

	var entries []CollabEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		entry := CollabEntry{
			Username: parts[0],
		}
		if len(parts) >= 2 {
			entry.AccessLevel = parts[1]
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

type keyValue struct {
	key   string
	value string
}

func parseKeyValues(output string) []keyValue {
	var kvs []keyValue
	for _, line := range strings.Split(output, "\n") {
		key, value, ok := parseKeyValue(line)
		if ok {
			kvs = append(kvs, keyValue{key: key, value: value})
		}
	}
	return kvs
}

func parseKeyValue(line string) (string, string, bool) {
	idx := strings.Index(line, ": ")
	if idx < 0 {
		// Handle lines ending with ":" (like "Public keys:" or "Branches:")
		if strings.HasSuffix(strings.TrimSpace(line), ":") {
			key := strings.TrimSuffix(strings.TrimSpace(line), ":")
			return key, "", true
		}
		return "", "", false
	}
	key := strings.TrimSpace(line[:idx])
	value := strings.TrimSpace(line[idx+2:])
	return key, value, true
}
