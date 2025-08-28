# Contributing Guide

## Prerequisites

- Go 1.24+
- Linux development environment (required for namespace/filesystem operations)
- SSH access to development server (see [vscode-usage.md](vscode-usage.md))

## Getting Started

1. **Clone and build**
   ```bash
   cd container-runtime
   make build
   ```

2. **Test basic functionality**
   ```bash
   make run-dev
   ```

## Development Workflow

### Code Structure

- `cmd/mrunc/` - CLI entry point
- `internal/cli/` - Command handling (run, child)
- `internal/runtime/` - Core container operations (namespaces, filesystem, execution)
- `internal/container/` - Configuration parsing
- `pkg/specs/` - Data structures and interfaces
- `configs/examples/` - Sample container configurations

### Making Changes

1. **Work in feature branches**
   ```bash
   git checkout -b feature/your-feature
   ```

2. **Test locally**
   ```bash
   make run-custom CONFIG=path/to/your/config.json
   ```

3. **Verify CI passes**
   - GitHub Actions runs on push to main
   - Build verification only (no functional tests yet)

### Container Configuration

Containers are defined in JSON files following this structure:
```json
{
  "root": {
    "path": "/path/to/rootfs",
    "readonly": false
  },
  "process": {
    "args": ["/bin/bash"],
    "env": ["PATH=/bin:/usr/bin"],
    "cwd": "/",
    "terminal": true
  },
  "hostname": "container-hostname"
}
```

### Testing Changes

- Use `make run-dev` for quick iteration
- Test with different rootfs paths and process configurations
- Verify namespace isolation works correctly
- Check filesystem operations don't break the host system

### Code Guidelines

- Follow Go conventions
- Add error handling for system calls
- Document any new configuration options
- Keep security implications in mind (this runs as root)

### Debugging

- Use `make run-custom` with debug configurations
- Check system logs if container creation fails
- Verify rootfs paths exist and are accessible
- Test on the provisioned development server to avoid breaking your local system

### Pull Requests

- Include clear description of changes
- Test with multiple container configurations
- Verify CI passes
- Update documentation if adding new features:
  - Add usage examples to `documentation/makefile-usage.md` for new make targets
  - Document configuration options in this file or create new docs in `documentation/`
  - Update `README.md` if changing core functionality
  - Add inline code comments for complex system calls or algorithms

### Common Issues

- **Permission denied**: Ensure running with `sudo`
- **Rootfs not found**: Verify paths in configuration files
- **Namespace creation fails**: Check kernel support for required namespaces
- **Pivot root fails**: Ensure target directories exist and have correct permissions

