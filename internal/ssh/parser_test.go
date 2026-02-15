package ssh

import (
	"testing"
)

func TestParseRepoInfo(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    RepoInfoResult
		wantErr bool
	}{
		{
			name: "full repo info",
			input: `Project Name: myproject
Repository: myrepo
Description: A test repository
Private: true
Hidden: false
Mirror: false
Owner: admin
Default Branch: main
Branches:
  - main
Tags:`,
			want: RepoInfoResult{
				ProjectName: "myproject",
				Repository:  "myrepo",
				Description: "A test repository",
				Private:     true,
				Hidden:      false,
				Mirror:      false,
				Owner:       "admin",
			},
		},
		{
			name: "minimal repo info",
			input: `Repository: bare-repo
Private: false
Hidden: false
Mirror: false
Default Branch: main`,
			want: RepoInfoResult{
				Repository: "bare-repo",
			},
		},
		{
			name: "repo with empty description",
			input: `Project Name:
Repository: test
Description:
Private: false
Hidden: true
Mirror: false`,
			want: RepoInfoResult{
				Repository: "test",
				Hidden:     true,
			},
		},
		{
			name:    "empty output",
			input:   "",
			wantErr: true,
		},
		{
			name:    "garbage input",
			input:   "this is not valid output",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRepoInfo(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseRepoInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.ProjectName != tt.want.ProjectName {
				t.Errorf("ProjectName = %q, want %q", got.ProjectName, tt.want.ProjectName)
			}
			if got.Repository != tt.want.Repository {
				t.Errorf("Repository = %q, want %q", got.Repository, tt.want.Repository)
			}
			if got.Description != tt.want.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.want.Description)
			}
			if got.Private != tt.want.Private {
				t.Errorf("Private = %v, want %v", got.Private, tt.want.Private)
			}
			if got.Hidden != tt.want.Hidden {
				t.Errorf("Hidden = %v, want %v", got.Hidden, tt.want.Hidden)
			}
			if got.Mirror != tt.want.Mirror {
				t.Errorf("Mirror = %v, want %v", got.Mirror, tt.want.Mirror)
			}
			if got.Owner != tt.want.Owner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.want.Owner)
			}
		})
	}
}

func TestParseUserInfo(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    UserInfoResult
		wantErr bool
	}{
		{
			name: "user with keys",
			input: `Username: alice
Admin: false
Public keys:
  ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA alice@laptop
  ssh-rsa AAAAB3NzaC1yc2EAAAA alice@desktop`,
			want: UserInfoResult{
				Username: "alice",
				Admin:    false,
				PublicKeys: []string{
					"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA alice@laptop",
					"ssh-rsa AAAAB3NzaC1yc2EAAAA alice@desktop",
				},
			},
		},
		{
			name: "admin user no keys",
			input: `Username: admin
Admin: true
Public keys:`,
			want: UserInfoResult{
				Username: "admin",
				Admin:    true,
			},
		},
		{
			name: "user with single key",
			input: `Username: bob
Admin: false
Public keys:
  ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA bob@host`,
			want: UserInfoResult{
				Username: "bob",
				Admin:    false,
				PublicKeys: []string{
					"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA bob@host",
				},
			},
		},
		{
			name:    "empty output",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseUserInfo(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseUserInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.Username != tt.want.Username {
				t.Errorf("Username = %q, want %q", got.Username, tt.want.Username)
			}
			if got.Admin != tt.want.Admin {
				t.Errorf("Admin = %v, want %v", got.Admin, tt.want.Admin)
			}
			if len(got.PublicKeys) != len(tt.want.PublicKeys) {
				t.Fatalf("PublicKeys length = %d, want %d", len(got.PublicKeys), len(tt.want.PublicKeys))
			}
			for i, key := range got.PublicKeys {
				if key != tt.want.PublicKeys[i] {
					t.Errorf("PublicKeys[%d] = %q, want %q", i, key, tt.want.PublicKeys[i])
				}
			}
		})
	}
}

func TestParseCollabList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []CollabEntry
	}{
		{
			name: "multiple collaborators",
			input: `alice read-write
bob read-only
charlie admin-access`,
			want: []CollabEntry{
				{Username: "alice", AccessLevel: "read-write"},
				{Username: "bob", AccessLevel: "read-only"},
				{Username: "charlie", AccessLevel: "admin-access"},
			},
		},
		{
			name:  "empty output",
			input: "",
			want:  nil,
		},
		{
			name:  "single collaborator",
			input: "alice read-write",
			want: []CollabEntry{
				{Username: "alice", AccessLevel: "read-write"},
			},
		},
		{
			name:  "username only without access level",
			input: "ssoriche",
			want: []CollabEntry{
				{Username: "ssoriche", AccessLevel: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCollabList(tt.input)
			if err != nil {
				t.Fatalf("ParseCollabList() error = %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("length = %d, want %d", len(got), len(tt.want))
			}
			for i, entry := range got {
				if entry.Username != tt.want[i].Username {
					t.Errorf("[%d] Username = %q, want %q", i, entry.Username, tt.want[i].Username)
				}
				if entry.AccessLevel != tt.want[i].AccessLevel {
					t.Errorf("[%d] AccessLevel = %q, want %q", i, entry.AccessLevel, tt.want[i].AccessLevel)
				}
			}
		})
	}
}
