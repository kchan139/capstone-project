# Makefile Usage

The `container-runtime/Makefile` provides common build and run commands for `mrunc`.

## Build

```bash
make build
```
→ Compiles the binary into `bin/mrunc`.

## Run

```bash
make run
```
→ Runs the built binary with default container name `default` and config `configs/examples/ubuntu.json`.

### Development Mode

```bash
make run-dev
```
→ Runs with `go run` using container name `dev`, useful for quick testing.

### Ubuntu Container
```bash
make run-ubuntu
```
→ Runs with `go run` using container name `ubuntu`.

### Custom Config

```bash
make run-custom NAME=mycontainer CONFIG=path/to/config.json
```
→ Runs with custom container name and config.

### Run Built Binary with Custom Config
```bash
make run-bin NAME=mycontainer CONFIG=path/to/config.json
```

## Create Container
```bash
make create NAME=mycontainer [CONFIG=path/to/config.json]
```
→ Creates a named container. Uses default config if not specified.

## Clean
```bash
make clean
```

→ Removes build artifacts.

## Install

```bash
make install
```

→ Installs `mrunc` to your local Go bin directory.

---
> TL;DR: Use `make build && make run` for production testing, or `make run-dev` for quick development runs. All run commands follow the pattern: `mrunc run <container-name> <config.json>`
