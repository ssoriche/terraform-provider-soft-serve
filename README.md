# Terraform Provider for Soft Serve

A Terraform provider for managing [Charm Soft Serve](https://github.com/charmbracelet/soft-serve) git server resources via SSH.

[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/ssoriche/terraform-provider-soft-serve)](https://goreportcard.com/report/github.com/ssoriche/terraform-provider-soft-serve)

## Features

- **Repositories** - Create and manage git repositories with visibility and description settings
- **Users** - Manage user accounts with SSH public key authentication
- **Repository Collaborators** - Configure per-repository access control for users
- **Server Settings** - Manage server-wide configuration like anonymous access and keyless auth

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for development)
- A running [Soft Serve](https://github.com/charmbracelet/soft-serve) instance

## Installation

### Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    softserve = {
      source  = "ssoriche/soft-serve"
      version = "~> 0.1"
    }
  }
}

provider "softserve" {
  host     = "localhost"
  port     = 23231
  username = "admin"
  use_agent = true
}
```

### Local Development

```bash
git clone https://github.com/ssoriche/terraform-provider-soft-serve
cd terraform-provider-soft-serve
make install
```

## Usage Examples

### Provider Configuration

```hcl
provider "softserve" {
  host     = "localhost"
  port     = 23231
  username = "admin"
  # private_key_path = "~/.ssh/id_ed25519"
  # identity_file    = "~/.ssh/id_ed25519.pub"
  use_agent = true
}
```

### User Management

```hcl
resource "softserve_user" "example" {
  username = "alice"
  admin    = false
  public_keys = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... alice@laptop",
  ]
}
```

### Repository Management

```hcl
resource "softserve_repository" "example" {
  name        = "my-project"
  description = "An example repository"
  private     = true
  hidden      = false
}
```

### Repository Collaborator

```hcl
resource "softserve_repository_collaborator" "example" {
  repository   = softserve_repository.example.name
  username     = softserve_user.example.username
  access_level = "read-write"
}
```

### Server Settings

```hcl
resource "softserve_server_settings" "this" {
  allow_keyless = false
  anon_access   = "read-only"
}
```

## Provider Configuration

### Arguments

- `host` - (Required) Soft Serve server hostname. Env: `SOFT_SERVE_HOST`
- `port` - (Optional) SSH port. Default: `23231`. Env: `SOFT_SERVE_PORT`
- `username` - (Optional) SSH username. Default: `admin`. Env: `SOFT_SERVE_USERNAME`
- `private_key_path` - (Optional) Path to SSH private key. Env: `SOFT_SERVE_PRIVATE_KEY_PATH`
- `identity_file` - (Optional) Path to SSH identity file. Env: `SOFT_SERVE_IDENTITY_FILE`
- `use_agent` - (Optional) Use SSH agent for authentication. Default: `false`. Env: `SOFT_SERVE_USE_AGENT`

### Environment Variables

```bash
export SOFT_SERVE_HOST="localhost"
export SOFT_SERVE_PORT="23231"
export SOFT_SERVE_USERNAME="admin"
export SOFT_SERVE_PRIVATE_KEY_PATH="~/.ssh/id_ed25519"
export SOFT_SERVE_USE_AGENT="true"
```

## Resources

- `softserve_user` - User accounts with SSH public key management
- `softserve_repository` - Git repositories with visibility settings
- `softserve_repository_collaborator` - Per-repository user access control
- `softserve_server_settings` - Server-wide configuration

## Development

### Building

```bash
make build
```

### Testing

```bash
# Unit tests
make test

# Acceptance tests (requires running Soft Serve instance)
make testacc
```

### Code Quality

```bash
# Format code
make fmt

# Lint
make lint
```

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes using conventional commits
4. Ensure all tests pass and code is formatted
5. Push to your branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

### Commit Message Format

This project follows [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(resource): add new repository resource
fix(ssh): handle connection timeout errors
docs(readme): update usage examples
chore(deps): update terraform-plugin-framework to v1.17.0
```

## Architecture

This provider is built using:

- [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) - Modern provider SDK
- SSH client - Communicates with Soft Serve via SSH commands
- Go 1.24 - Implementation language

### Project Structure

```
terraform-provider-soft-serve/
├── internal/
│   ├── ssh/             # SSH client and output parser
│   │   ├── client.go    # SSH connection and command execution
│   │   ├── parser.go    # Soft Serve output parsing
│   │   └── parser_test.go
│   ├── provider/        # Terraform provider configuration
│   │   └── provider.go
│   └── resource/        # Terraform resources
│       ├── repository.go
│       ├── repository_collaborator.go
│       ├── server_settings.go
│       └── user.go
├── examples/            # Usage examples
│   ├── provider/
│   └── resources/
├── devbox.json          # Development environment
└── main.go              # Provider entry point
```

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.
