# Makefile usage

Run all commands below from `container-runtime/`.

## Build

```bash
make version
make build
make build-race
```

- `version` prints build metadata.
- `build` writes `bin/mrunc`.
- `build-race` writes `bin/mrunc-race`.

## Test

```bash
make test
make test-unit
make test-integration
make test-coverage
make test-all
```

## Maintenance

```bash
make clean
make install
make dev-setup
make help
```

## Runtime flow

Build the binary first:

```bash
make build
```

Initialize the default Ubuntu rootfs and config:

```bash
sudo ./bin/mrunc init
```

Run from a bundle directory that contains `config.json`:

```bash
sudo ./bin/mrunc run --bundle /path/to/bundle <container-id>
```

Create and start separately:

```bash
sudo ./bin/mrunc create --bundle /path/to/bundle <container-id>
sudo ./bin/mrunc start <container-id>
```

Inspect and clean up:

```bash
./bin/mrunc list
sudo ./bin/mrunc kill <container-id>
sudo ./bin/mrunc delete <container-id>
```

## Toolchain check

`make build` and `make test-unit` call `check-go-version`.

```bash
go env GOVERSION
# go1.25.8
```
