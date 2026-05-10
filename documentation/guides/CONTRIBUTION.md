# Contributing

## Prerequisites

- Linux.
- Go `1.25.10`.
- `make`.
- `libseccomp` development headers for builds with seccomp support.
- Root or `sudo` for runtime commands that create namespaces, mounts, cgroups, and network devices.

The Makefile enforces `go1.25.10` with `check-go-version`.

## Build and test

```bash
cd container-runtime
make build
make test-unit
make test-integration
```

Other useful targets:

```bash
make build-race
make test-coverage
make test-all
make version
make clean
make install
make dev-setup
```

## Direct CLI workflow

`mrunc` reads `config.json` from the bundle directory passed with `--bundle`.

Build the binary:

```bash
cd container-runtime
make build
```

Initialize the default Ubuntu rootfs and default config:

```bash
sudo ./bin/mrunc init
```

Run a container from a bundle:

```bash
sudo ./bin/mrunc run --bundle /path/to/bundle <container-id>
```

Create and start separately:

```bash
sudo ./bin/mrunc create --bundle /path/to/bundle <container-id>
sudo ./bin/mrunc start <container-id>
```

List, signal, and remove containers:

```bash
./bin/mrunc list
sudo ./bin/mrunc kill <container-id>
sudo ./bin/mrunc delete <container-id>
./bin/mrunc version
```

## Code layout

- `cmd/mrunc/` — entry point.
- `internal/cli/` — user-facing commands and internal helper commands.
- `internal/container/` — config loading and validation.
- `internal/runtime/` — namespaces, cgroups, mounts, devices, pty, seccomp, capabilities, networking, and state updates.
- `internal/utils/` — path, env, pipe, and socket helpers.
- `configs/examples/` — example bundle configs.
- `tests/` — integration tests and test helpers.

## Documentation

When runtime behavior changes, update the matching file under `documentation/` in the same change.
