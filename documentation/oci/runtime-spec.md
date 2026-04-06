# Runtime reference

## Commands

The current CLI provides these user-facing commands:

| Command | Purpose |
|---|---|
| `run` | Create and run a container from a bundle. |
| `create` | Create a container and wait for `start`. |
| `start` | Start a container that was created earlier. |
| `list` | List containers and their metadata. |
| `kill` | Signal a container process or all processes in its cgroup. |
| `delete` | Remove container runtime state and related resources. |
| `init` | Download the default Ubuntu rootfs and example config. |
| `version` | Print version, commit, and build date. |

The runtime also uses these internal commands:

- `child`
- `intermediate`
- `initproc`
- `monitor`

## Bundle handling

`run` and `create` accept `--bundle` and load `config.json` from that directory.

## Runtime state

The runtime stores container state under `/run/mrunc/<container-id>/`.

`start` signals the created container by opening `/run/mrunc/<container-id>/exec.fifo` and writing one byte.

## Namespaces

The runtime maps `linux.namespaces` into `syscall.SysProcAttr` clone flags for:

- `pid`
- `network`
- `ipc`
- `uts`
- `mount`
- `cgroup`

## Filesystem setup

The container process path:

1. bind-mounts the rootfs,
2. applies configured mounts,
3. bind-mounts the cgroup path for `cgroup` and `cgroup2` mounts,
4. calls `pivot_root`,
5. changes working directory,
6. removes `/put_old`.

## Mount options

The runtime mount parser handles:

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

## Devices and terminal

When `/dev` is prepared, the runtime creates:

- `/dev/null`
- `/dev/zero`
- `/dev/full`
- `/dev/random`
- `/dev/urandom`
- `/dev/tty`

When `process.terminal` is true, the runtime sets up a pty and can send the console file descriptor through `--console-socket`.

## Cgroups

The runtime creates a cgroup v2 directory and writes:

- `cpu.weight`
- `cpu.max`
- `memory.max`
- `memory.low`
- `memory.swap.max`
- `pids.max`

## Networking

If `linux.network.enableNetwork` is true, the runtime can:

- create a veth pair,
- move one side into the container network namespace,
- bring up loopback,
- configure the container IP address,
- add the default route,
- write `/etc/resolv.conf`,
- run the configured firewall script,
- remove the host veth on exit.

`linux.network` is a project-specific extension used by the checked-in examples.

## Capabilities and seccomp

Before `exec`, the runtime can:

- apply capability sets from `process.capabilities`,
- set `PR_SET_NO_NEW_PRIVS` when `process.noNewPrivileges` is true,
- load a seccomp profile from `linux.seccomp`.

## `init`

`mrunc init` does two things:

1. downloads and extracts the Ubuntu 24.04 minimal rootfs into `/var/lib/mrunc/images/ubuntu`,
2. downloads the default `ubuntu.json` example config into that directory.
