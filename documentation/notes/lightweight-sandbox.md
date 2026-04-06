# Isolation and hardening

## Namespaces

`runtime.CreateNamespaces(config)` maps these namespace types into clone flags:

- `pid`
- `network`
- `ipc`
- `uts`
- `mount`
- `cgroup`

## Filesystem isolation

The container process path in `internal/cli/child.go` and `internal/cli/initproc.go` does the following:

1. bind-mounts the configured rootfs onto itself,
2. applies configured mounts,
3. bind-mounts `config.CgroupPath` for `cgroup` and `cgroup2` mounts,
4. calls `pivot_root`,
5. changes working directory,
6. unmounts and removes `/put_old`.

## Device setup

When `/dev` is prepared, `runtime.SetupDev` creates these character devices:

- `null`
- `zero`
- `full`
- `random`
- `urandom`
- `tty`

It also creates these symlinks:

- `/dev/fd`
- `/dev/stdin`
- `/dev/stdout`
- `/dev/stderr`

When a terminal is requested, the runtime also links `/dev/ptmx` to `pts/ptmx` and can bind the pty slave to `/dev/console`.

## Cgroups

`runtime.CreateCgroup` writes cgroup v2 files for:

- `cpu.weight`
- `cpu.max`
- `memory.max`
- `memory.low`
- `memory.swap.max`
- `pids.max`

The cgroup is created under the parent of the current process cgroup path.

## Networking

If `linux.network.enableNetwork` is true, the runtime can:

- create a host/container veth pair,
- move the container-side veth into the container network namespace,
- bring up loopback inside the container,
- assign the container address,
- add the default route,
- write `/etc/resolv.conf`,
- run the configured firewall script,
- remove the host veth on exit.

## Capabilities

`runtime.SetupCaps` applies capability settings from `process.capabilities`:

- `bounding`
- `permitted`
- `effective`
- `inheritable`
- `ambient`

## `no_new_privs`

If `process.noNewPrivileges` is true, the runtime calls `prctl(PR_SET_NO_NEW_PRIVS, 1, ...)` before `exec`.

## Seccomp

`runtime.SetupSeccomp` loads a seccomp filter from `linux.seccomp`:

- `defaultAction`
- `architectures`
- `syscalls[].names`
- `syscalls[].action`

## Fanotify monitor

The optional `--fanotify-monitor` path uses watch rules with:

- `path`
- `events`
- `action`
