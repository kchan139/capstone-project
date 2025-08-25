# Makefile Usage

The `container-runtime/Makefile` provides common build and run commands for `mrunc`.

## Build

```bash
make build
````

→ Compiles the binary into `bin/mrunc`.

## Run

```bash
make run
```

→ Runs the built binary with the default config (`configs/examples/ubuntu.json`).

### Development Mode

```bash
make run-dev
```

→ Runs with `go run`, useful for quick testing.

### Custom Config

```bash
make run-custom CONFIG=path/to/config.json
```

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

> TL;DR: Use `make build && make run` for production testing, or `make run-dev` for quick development runs.
