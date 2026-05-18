# PPE Quick Start Guide for logsift

This guide provides quick instructions for setting up and using the Prompt Processing Environment (PPE) for the logsift project.

## Prerequisites

- Docker installed and running
- Git installed
- SSH client (for Trae connection)

## Quick Setup

### 1. Build the Trae SSH Docker Image

```bash
# From the project root
docker build -f environment/Dockerfile -t logsift-trae .
```

### 2. Start the Development Container

```bash
docker run -d \
  --name logsift-trae \
  -p 2222:22 \
  -e SSH_PASSWORD=password \
  logsift-trae
```

### 3. Configure SSH (One-time setup)

Add to your `~/.ssh/config`:

```sshconfig
Host logsift-trae
    HostName localhost
    User root
    Port 2222
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
```

### 4. Connect with Trae

Use Trae's SSH remote development feature to connect to `logsift-trae`.

## Essential Commands

### Build and Test

```bash
# Build the binary
make build

# Run tests
make test

# Format code
make fmt

# Static analysis
make vet
```

### Run the Application

```bash
# Build first
make build

# Run with sample data
./logsift --file testdata/sample.ndjson --output tsv

# Run with TUI interface
./logsift --file testdata/sample.ndjson --tui
```

### Docker Operations

```bash
# Build standard Docker image
make docker

# Run tests in container
docker run --rm -it logsift bash -c 'cd /app && go test ./...'

# Enter container shell
make docker-shell
```

## Development Workflow

1. **Setup Environment**
   ```bash
   # Connect via Trae SSH
   # Container is already at /app with all dependencies
   ```

2. **Make Changes**
   ```bash
   # Edit files in /app
   # Use your preferred editor via Trae
   ```

3. **Test Changes**
   ```bash
   make test
   make vet
   ```

4. **Build and Run**
   ```bash
   make build
   ./logsift --file testdata/sample.ndjson
   ```

## Common Tasks

### Adding Dependencies

```bash
# Add a new dependency
go get github.com/example/new-package

# Update go.mod and go.sum
go mod tidy
```

### Running Specific Tests

```bash
# Run tests for a specific package
go test ./internal/parser/

# Run with verbose output
go test -v ./internal/filter/

# Run with coverage
go test -cover ./...
```

### Generating Submission Package

```bash
# Create submission package for grading
make submission
# Output: submissions/logsift/Dockerfile + submissions/logsift/repo.zip
```

## Troubleshooting

### SSH Connection Issues

```bash
# Check if container is running
docker ps | grep logsift-trae

# Check container logs
docker logs logsift-trae

# Restart container
docker restart logsift-trae
```

### Go Command Not Found

```bash
# Use bash -c instead of bash -lc
# Correct:
docker run --rm -it logsift bash -c 'cd /app && go version'

# Incorrect (PATH will be stripped):
docker run --rm -it logsift bash -lc 'cd /app && go version'
```

### Network/Proxy Issues

```bash
# Build with proxy settings
docker build \
  --build-arg http_proxy=http://your-proxy:port \
  --build-arg https_proxy=http://your-proxy:port \
  -f environment/Dockerfile \
  -t logsift-trae .
```

## Environment Details

- **Working Directory**: `/app`
- **Go Version**: 1.24.0
- **Base Image**: `golang:1.24-bookworm`
- **SSH Access**: Port 2222, user `root`, password `password`
- **Key Tools**: git, curl, tmux, openssh-server

## Quick Reference

| Command | Description |
|---------|-------------|
| `make build` | Build logsift binary |
| `make test` | Run all tests |
| `make docker` | Build Docker image |
| `make submission` | Generate submission package |
| `./logsift --file testdata/sample.ndjson` | Run with sample data |
| `./logsift --tui` | Run with TUI interface |

## Getting Help

- Check `README.md` for detailed project documentation
- Review `docs/TRAE_DOCKER.md` for Trae-specific setup
- See `.trae/ppe_config.yaml` for full PPE configuration
