# kubectl-auth-vault

[![CI](https://github.com/efortin/kubectl-auth-vault/actions/workflows/ci.yml/badge.svg)](https://github.com/efortin/kubectl-auth-vault/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/efortin/kubectl-auth-vault)](https://goreportcard.com/report/github.com/efortin/kubectl-auth-vault)

A kubectl credential plugin that fetches OIDC tokens from HashiCorp Vault for Kubernetes authentication.

## Features

- Fetches OIDC tokens from Vault using the official [vault-client-go](https://github.com/hashicorp/vault-client-go) SDK
- Caches tokens locally until expiration
- Configurable Vault address and token path
- Cross-platform support (Linux, macOS, Windows)
- Works as a kubectl plugin (`kubectl auth-vault`)

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap efortin/tap
brew install kubectl-auth_vault
```

### From releases

Download the latest release from the [releases page](https://github.com/efortin/kubectl-auth-vault/releases).

### From source

```bash
# Install task runner if not already installed
brew install go-task

# Build for your platform
task build

# Install to /usr/local/bin
task install

# Or install to ~/bin
task install:user
```

### Build all platforms

```bash
task build:all
```

Binaries will be in `dist/` directory.

## Usage

### Commands

```bash
# Get token (main command for kubeconfig exec)
kubectl auth-vault get --vault-addr https://vault.example.com --token-path identity/oidc/token/my_role

# Test configuration
kubectl auth-vault config test --vault-addr https://vault.example.com

# Show current configuration
kubectl auth-vault config show

# Show version
kubectl auth-vault version
```

### Command line

```bash
# Using environment variable
export VAULT_ADDR=https://vault.example.com
kubectl-auth_vault get --token-path identity/oidc/token/kubernetes

# Using flags
kubectl-auth_vault get --vault-addr https://vault.example.com --token-path identity/oidc/token/my_role

# Disable caching
kubectl-auth_vault get --token-path identity/oidc/token/my_role --no-cache
```

### Kubeconfig integration

Add the following to your `~/.kube/config`:

```yaml
users:
- name: vault-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1
      command: kubectl-auth_vault
      interactiveMode: Never
      env:
      - name: VAULT_ADDR
        value: https://vault.example.com
      args:
      - get
      - --token-path
      - identity/oidc/token/kubernetes
```

Or with inline vault-addr:

```yaml
users:
- name: vault-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1
      command: kubectl-auth_vault
      interactiveMode: Never
      args:
      - get
      - --vault-addr
      - https://vault.<your-domain>
      - --token-path
      - identity/oidc/token/kubernetes
```

## Get Command Options

| Flag | Environment | Description | Default |
|------|-------------|-------------|---------|
| `--vault-addr` | `VAULT_ADDR` | Vault server address | (required) |
| `--token-path` | - | Vault OIDC token path | `identity/oidc/token/kubernetes` |
| `--cache-file` | - | Token cache file path | `~/.kube/vault_<path>_token.json` |
| `--no-cache` | - | Disable token caching | `false` |

## Authentication with Vault

The plugin uses standard Vault authentication methods via environment variables:

- `VAULT_TOKEN` - Direct token authentication
- `VAULT_ROLE_ID` / `VAULT_SECRET_ID` - AppRole authentication
- Vault agent socket if running

Make sure you are authenticated to Vault before using the plugin:

```bash
vault login -method=oidc
```

## Development

```bash
# Download dependencies
task deps

# Run tests
task test

# Run tests with coverage
task test:coverage

# Run linter
task lint

# Build
task build

# Build release locally
task release:local

# Clean
task clean
```

## Project Structure

```
kubectl-auth-vault/
├── cmd/
│   └── kubectl-auth_vault/    # Main entry point
├── internal/
│   ├── cmd/                   # CLI commands (Cobra)
│   ├── vault/                 # Vault client wrapper
│   ├── cache/                 # Token caching
│   ├── credential/            # ExecCredential output
│   └── jwt/                   # JWT parsing utilities
├── .github/workflows/         # CI/CD workflows
├── .goreleaser.yml            # GoReleaser configuration
└── Taskfile.yml               # Task runner configuration
```

## License

MIT
