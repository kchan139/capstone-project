# `config.json` shape used by `mrunc`

`mrunc` reads `config.json` from the bundle directory passed with `--bundle`.

The checked-in examples and runtime code use an OCI-style layout with a project-specific network section under `linux.network`.

## Top-level fields

- `root`
- `process`
- `hostname`
- `mounts`
- `linux`

## `root`

The examples use:

- `path`
- `readonly`

## `process`

The runtime uses:

- `args`
- `env`
- `cwd`
- `terminal`
- `user.uid`
- `user.gid`
- `user.additionalGids`
- `noNewPrivileges`
- `capabilities.bounding`
- `capabilities.permitted`
- `capabilities.effective`
- `capabilities.inheritable`
- `capabilities.ambient`

## `mounts`

Each mount entry uses:

- `destination`
- `type`
- `source`
- `options`

The mount option parser handles:

- `nosuid`
- `noexec`
- `nodev`
- `ro`
- `readonly`
- `bind`
- `rbind`
- `relatime`
- `noatime`
- `strictatime`

`cgroup` and `cgroup2` mounts are handled by bind-mounting `config.CgroupPath`.

## `linux.namespaces`

The runtime handles these namespace `type` values:

- `pid`
- `network`
- `ipc`
- `uts`
- `mount`
- `cgroup`

## `linux.resources`

The runtime uses these resource fields:

- `cpu.shares`
- `cpu.quota`
- `cpu.period`
- `memory.limit`
- `memory.reservation`
- `memory.swap`
- `pids.limit`

## `linux.network`

The project-specific network section uses:

- `enableNetwork`
- `containerIP`
- `gatewayCIDR`
- `vethHost`
- `vethContainer`
- `dns`
- `firewallScript`

## `linux.seccomp`

The runtime uses:

- `defaultAction`
- `architectures`
- `syscalls[].names`
- `syscalls[].action`

## Fanotify monitor config

The optional `--fanotify-monitor` config uses:

- `watch_rules[]`
- `path`
- `events`
- `action`
