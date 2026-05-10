# MRUNC Makefile Usage

This document describes the supported `make` targets for building, testing, running, and managing MRUNC containers.

Run all commands from:

```bash
cd container-runtime
````

## Variables

| Variable              |                          Default | Purpose                                                |
| --------------------- | -------------------------------: | ------------------------------------------------------ |
| `NAME`                |                        `default` | Container ID used by lifecycle commands                |
| `CONFIG`              | `./configs/examples/ubuntu.json` | Container config file copied into the temporary bundle |
| `BUNDLE_DIR`          |      `/tmp/mrunc-default-bundle` | Temporary OCI-style bundle directory                   |
| `SIGNAL`              |                           `TERM` | Signal used by `make kill`                             |
| `GO_REQUIRED_VERSION` |                       `go1.25.10` | Required Go version                                    |

Example override:

```bash
make run NAME=demo CONFIG=./configs/examples/ubuntu.json
```

## Build targets

### Build the binary

```bash
make build
```

Builds:

```text
./bin/mrunc
```

The build injects version metadata from Git:

* version
* commit
* build date

### Build with race detection

```bash
make build-race
```

Builds:

```text
./bin/mrunc-race
```

### Show version metadata

```bash
make version
```

### Clean build artifacts

```bash
make clean
```

Removes:

```text
./bin
coverage.out
coverage.html
```

## Run targets

The runtime expects an OCI-style bundle directory containing `config.json`.

The Makefile creates a temporary bundle by copying the selected config file into:

```text
$(BUNDLE_DIR)/config.json
```

Then it calls MRUNC with:

```bash
mrunc run --bundle "$(BUNDLE_DIR)" "$(NAME)"
```

### Run built binary

```bash
make run
```

Equivalent behavior:

```bash
make run NAME=default CONFIG=./configs/examples/ubuntu.json
```

Custom container name:

```bash
make run NAME=demo
```

Custom config:

```bash
make run NAME=demo CONFIG=./configs/examples/ci-test.json
```

### Run through `go run`

```bash
make run-dev
```

Uses container ID:

```text
dev
```

### Run Ubuntu example through `go run`

```bash
make run-ubuntu
```

Uses container ID:

```text
ubuntu
```

## Container lifecycle targets

Use these when testing the separated lifecycle flow:

```text
create -> start -> kill -> delete
```

### Create container

```bash
make create NAME=demo
```

With custom config:

```bash
make create NAME=demo CONFIG=./configs/examples/ci-test.json
```

This creates a container from the generated bundle but does not start the container process.

### Start created container

```bash
make start NAME=demo
```

### List containers

```bash
make list
```

### Send signal to container init process

Default signal is `TERM`:

```bash
make kill NAME=demo
```

Use a specific signal:

```bash
make kill NAME=demo SIGNAL=KILL
make kill NAME=demo SIGNAL=9
make kill NAME=demo SIGNAL=INT
```

### Kill all processes in the container cgroup

```bash
make kill-all NAME=demo
```

Use this when the container has child processes that may survive after the init process exits.

### Delete stopped container

```bash
make delete NAME=demo
```

This expects the container to already be stopped.

### Force-delete container

```bash
make delete-force NAME=demo
```

Use this for stuck or still-running containers.

### Best-effort cleanup

```bash
make clean-container NAME=demo
```

This runs:

```bash
mrunc kill --all demo
mrunc delete --force demo
```

Both commands are prefixed with `-`, so cleanup continues even if one step fails.

### Create, start, and list

```bash
make lifecycle NAME=demo
```

This runs:

```text
create -> start -> list
```

## Test targets

### Unit tests

```bash
make test
```

or:

```bash
make test-unit
```

### Integration tests

```bash
make test-integration
```

### All tests

```bash
make test-all
```

### Coverage report

```bash
make test-coverage
```

Generates:

```text
coverage.out
coverage.html
```

## Development setup

```bash
make dev-setup
```

Runs:

```bash
go mod download
go mod tidy
```

## Install binary

```bash
make install
```

Installs the binary through:

```bash
go install ./cmd/mrunc
```

## Common workflows

### One-shot run

```bash
make run NAME=demo
```

### Two-phase lifecycle

```bash
make create NAME=demo
make start NAME=demo
make list
make kill NAME=demo SIGNAL=TERM
make delete NAME=demo
```

### Hard cleanup

```bash
make clean-container NAME=demo
```

### Recreate a container cleanly

```bash
make clean-container NAME=demo
make create NAME=demo
make start NAME=demo
```

## Notes

`run` and `create` use `--bundle`, not a direct positional config path.

Correct:

```bash
sudo ./bin/mrunc run --bundle /tmp/mrunc-default-bundle demo
```

Wrong:

```bash
sudo ./bin/mrunc run demo ./configs/examples/ubuntu.json
```

The second form is stale. It causes MRUNC to resolve the config as:

```text
./config.json
```

from the current working directory, which fails unless that file exists.

`kill` sends the requested signal, but process shutdown may not be immediate. If `delete` fails after `SIGTERM`, either wait and retry or use:

```bash
make delete-force NAME=demo
```
