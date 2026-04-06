# What “minimal” means in `mrunc`

In this project, “minimal” is about keeping the runtime focused on container startup and teardown.

## Runtime steps

The runtime handles these core steps:

- load `config.json`,
- create namespaces,
- create a cgroup,
- prepare mounts and call `pivot_root`,
- prepare `/dev` and an optional terminal,
- configure optional veth-based networking,
- apply capabilities,
- apply seccomp,
- `exec` the configured process,
- write runtime state under `/run/mrunc/<id>/`.

## User-facing commands

The current CLI exposes:

- `run`
- `init`
- `version`
- `create`
- `start`
- `list`
- `delete`
- `kill`

The runtime also uses these internal commands:

- `child`
- `intermediate`
- `initproc`
- `monitor`

## Project shape

Most runtime logic lives in a small set of packages:

- `internal/cli/`
- `internal/container/`
- `internal/runtime/`
- `internal/utils/`

The checked-in example configs under `configs/examples/` are part of that small surface area.
