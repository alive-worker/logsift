# Project Rules for logsift

This file defines the project-specific rules and commands for the logsift project.

## Build and Development Commands

### Build Commands
```bash
# Build the binary
go build -o logsift ./cmd/logsift

# Or using Makefile
make build
```

### Test Commands
```bash
# Run all tests
go test ./...

# Or using Makefile
make test

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Code Quality Commands
```bash
# Format Go code
gofmt -w .

# Or using Makefile
make fmt

# Run static analysis
go vet ./...

# Or using Makefile
make vet
```

### Docker Commands
```bash
# Build standard Docker image
docker build -t logsift .

# Or using Makefile
make docker

# Build Trae SSH Docker image
docker build -f environment/Dockerfile -t logsift-trae .

# Run Trae SSH container
docker run -d --name logsift-trae -p 2222:22 -e SSH_PASSWORD=password logsift-trae
```

## Code Quality Standards

### Before Committing Code
Always run these checks before committing:

1. **Format code**
   ```bash
   make fmt
   ```

2. **Run static analysis**
   ```bash
   make vet
   ```

3. **Run tests**
   ```bash
   make test
   ```

### Expected Test Coverage
- Minimum test coverage: 80%
- Run coverage check: `go test -cover ./...`

## Project Structure Conventions

### Source Code Organization
- Main entry point: `cmd/logsift/main.go`
- Internal packages: `internal/` directory
- Test files: `*_test.go` alongside source files
- Sample data: `testdata/` directory

### Naming Conventions
- Package names: lowercase, single word
- File names: snake_case for test files
- Function names: CamelCase for exported, camelCase for unexported
- Variable names: descriptive, camelCase

## Development Environment

### Required Tools
- Go 1.24 or later
- Git
- Docker (for containerized development)
- SSH client (for Trae connection)

### Environment Variables
```bash
# Required for Go module proxy in China
export GOPROXY=https://goproxy.cn,direct
export CGO_ENABLED=0

# Optional for proxy settings
export http_proxy=http://your-proxy:port
export https_proxy=http://your-proxy:port
```

### SSH Configuration for Trae
```sshconfig
Host logsift-trae
    HostName localhost
    User root
    Port 2222
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
```

## Troubleshooting

### Common Issues and Solutions

#### 1. Go Command Not Found in Docker
**Problem**: When running `docker run --rm -it logsift bash -lc 'go version'`, go command is not found.

**Solution**: Use `bash -c` instead of `bash -lc`:
```bash
docker run --rm -it logsift bash -c 'go version'
```

#### 2. Network Connectivity Issues
**Problem**: `go mod download` fails due to network issues.

**Solution**: Set proxy environment variables:
```bash
# For Docker build
docker build \
  --build-arg http_proxy=http://your-proxy:port \
  --build-arg https_proxy=http://your-proxy:port \
  -f environment/Dockerfile \
  -t logsift-trae .

# For container runtime
docker run -e http_proxy=http://your-proxy:port -e https_proxy=http://your-proxy:port ...
```

#### 3. SSH Connection Failed
**Problem**: Cannot connect to Trae SSH container.

**Solution**:
```bash
# Check if container is running
docker ps | grep logsift-trae

# Check container logs
docker logs logsift-trae

# Restart container
docker restart logsift-trae

# Verify SSH service is running inside container
docker exec logsift-trae ps aux | grep sshd
```

## Automation Scripts

### Pre-commit Hook Example
Create `.git/hooks/pre-commit`:
```bash
#!/bin/bash
echo "Running pre-commit checks..."

# Format code
make fmt

# Run static analysis
if ! make vet; then
    echo "go vet failed"
    exit 1
fi

# Run tests
if ! make test; then
    echo "Tests failed"
    exit 1
fi

echo "All checks passed!"
```

### CI/CD Pipeline Commands
```yaml
# Example GitHub Actions workflow
steps:
  - name: Checkout code
    uses: actions/checkout@v3
    
  - name: Set up Go
    uses: actions/setup-go@v4
    with:
      go-version: '1.24'
      
  - name: Run tests
    run: make test
    
  - name: Run static analysis
    run: make vet
    
  - name: Build binary
    run: make build
    
  - name: Build Docker image
    run: make docker
```

## Documentation

### Required Documentation Updates
When making significant changes, update:
1. `README.md` - Main project documentation
2. `docs/DEVELOPMENT_PLAN.md` - Development plan
3. `docs/ROADMAP.md` - Project roadmap
4. `.trae/ppe_config.yaml` - PPE configuration
5. `.trae/QUICK_START.md` - Quick start guide

### Code Comments
- Export functions must have GoDoc comments
- Complex algorithms should have inline comments
- TODO comments should include issue reference if available

## Security Guidelines

### Secrets Management
- Never commit secrets to repository
- Use environment variables for sensitive data
- Docker images should not contain hardcoded secrets

### Dependency Security
- Regularly update dependencies: `go get -u ./...`
- Check for vulnerabilities: `go list -m all | grep -E "(vulnerability|CVE)"`
- Use trusted package sources

## Performance Guidelines

### Memory Usage
- Use buffers for I/O operations
- Release resources promptly (defer close)
- Avoid memory leaks in long-running processes

### CPU Optimization
- Use efficient algorithms for log filtering
- Consider concurrency for parallel processing
- Profile performance: `go test -bench=. ./...`

## Compliance

### License Requirements
- All code must comply with project license
- Third-party dependencies must have compatible licenses
- License headers must be included in source files

### Code Review Requirements
- All changes must be reviewed before merging
- Tests must pass before code review
- Documentation must be updated for new features
