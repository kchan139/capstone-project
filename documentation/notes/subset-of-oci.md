###  **Required attributes in `config.json`**

#### **Top-level**

- `ociVersion` (string)
    
- `process` (object)
    
- `root` (object)
    
- `linux` (object)
    

#### **Inside `process`**

- `args` (array of strings) — command and its arguments
    
- `user.uid` and `user.gid` — container user mapping
    

#### **Inside `root`**

- `path` (string) — path to root filesystem directory
    

#### **Inside `linux`**

- `namespaces` (array) — must include at least: `mount`, `pid`, `uts`, `user`
    
- `uidMappings` and `gidMappings` — required when user namespace is enabled
    
- `resources` (object) — must define at least:
    
    - `pids.max`
        
    - `cpu.max` or `cpu.weight`
        
    - `memory.limit`
        

---

### **Implement later**

- `hostname`
    
- `env`, `cwd`, `terminal`
    
- `mounts` (limited types: `bind`, `proc`, `tmpfs`, `devpts`)
    
- `process.capabilities` (subset: `bounding`, `effective`, `inheritable`)
    
- `seccomp` (subset with `defaultAction`, `syscalls`)
    

---

###  **Ignored / Not supported**

- `annotations`, `hooks`, `solaris`, `windows`
    
- `apparmorProfile`, `selinuxLabel`
    
- `maskedPaths`, `readonlyPaths`
    
- `devices`, `cgroupsPath`
    
- network, ipc, time, cgroup namespaces
--------------------------------------


# MRUNC: OCI Runtime Subset (Linux-only)

## 0) Scope & assumptions

- **Platform:** Linux only.
    
- **Bundles:** Standard OCI bundle layout (`/path/to/bundle/{config.json, rootfs/}`).
    
- **Lifecycle commands:** `create`, `start`, `run`, `delete` (+ optional `state`).
    
- **Single-process model:** One entry process per container (no extra `exec` API).
    
- **No daemon**; MRUNC is a single binary.
    

---

## 1) Implemented OCI objects & fields

### 1.1 `config.json` (subset)

**Top-level**

- `ociVersion` (string) — parsed but not strictly enforced beyond major.minor compatibility.
    
- `process` (object) — **required**
    
- `root` (object) — **required**
    
- `hostname` (string) — **optional**
    
- `mounts` (array of mount objects) — **optional**
    
- `linux` (object) — **required** (subset described below)
    
- **Ignored if present:** `annotations`, `hooks`, `solaris`, `windows`
    

**`process`**

- `terminal` (bool) — optional (if `true`, MRUNC allocates a pty).
    
- `cwd` (string) — default `/`.
    
- `args` (array<string>) — **required** (argv0..).
    
- `env` (array<string>) — optional.
    
- `rlimits` — **not supported** (omit or ignored).
    
- `user` — interpreted **inside** the user namespace:
    
    - `uid` (uint32), `gid` (uint32)
        
    - `additionalGids` (array<uint32>) — optional
        
- `noNewPrivileges` (bool) — **always enforced true** (MRUNC sets it regardless of input).
    
- `capabilities` (object) — subset:
    
    - **Used to construct effective/perm/inheritable/bounding=desired set** but MRUNC will **start from empty** and only add what’s listed under `bounding` + `effective`. If field omitted → **no caps**.
        
    - Supported lists: `bounding`, `effective`, `inheritable`. (Others ignored.)
        
- `apparmorProfile`, `selinuxLabel` — **not supported** (ignored).
    

**`root`**

- `path` (string) — **required** (path to `rootfs/`).
    
- `readonly` (bool) — **default true** (MRUNC may remount ro even if omitted).
    

**`mounts`**

- Supported `type`: `bind`, `proc`, `tmpfs`, `devpts` (anything else → reject).
    
- `destination`, `source`, `options` supported.
    
- Default security options MRUNC adds if not present: `nosuid,nodev,noexec` (except where the FS inherently requires otherwise, e.g., devpts).
    
- MRUNC will refuse mounts that violate sandbox policy (e.g., bind mounts that cross out of allowed roots unless explicitly enabled).
    

**`linux`**

- `namespaces` (array):
    
    - Supported `type`: `"mount"`, `"pid"`, `"uts"`, `"user"`.
        
    - `path` ignored (MRUNC always creates new namespaces; no `setns` into existing).
        
- `maskedPaths`, `readonlyPaths` — **ignored** (MRUNC enforces via global mount policy instead).
    
- `uidMappings`, `gidMappings` — **required** if using non-root host; MRUNC requires a valid mapping when `user` ns is enabled (it always is).
    
- `resources` (subset; **cgroups v2 only**):
    
    - `pids.max` → limit number of processes.
        
    - `cpu` → `max` (throttle) and/or `weight` (if present) → mapped to cgroup v2.
        
    - `memory` → `limit` (bytes) mapped to `memory.max`; optional `swap` ignored.
        
    - `blockIO/io` → optional simple throttles if provided; otherwise ignored.
        
- `seccomp` (subset):
    
    - `defaultAction`: `"SCMP_ACT_KILL"` or `"SCMP_ACT_ERRNO"` (MRUNC default: **KILL**).
        
    - `architectures` — optional; if omitted MRUNC autodetects native arch.
        
    - `syscalls`: MRUNC expects an **allowlist** (each entry: `names`, `action="SCMP_ACT_ALLOW"`, optional `args` constraints).
        
    - If `seccomp` absent → MRUNC applies its **own conservative default allowlist**.
        
- `devices` / `cgroupsPath` — **ignored** (MRUNC creates a private `/dev` and a per-container cgroup path automatically).
    
- `mountLabel`, `rootfsPropagation` — **ignored**.
    

---

## 2) Lifecycle & state (subset)

- **`create`**: Validates bundle, prepares namespaces, cgroup, and filesystem; creates container state directory.
    
- **`start`**: Executes the user process in the prepared container.
    
- **`run`**: `create` + `start` in one step.
    
- **`delete`**: Removes state dir, cgroup, any temp mounts.
    
- **`state`** (optional): Returns `{ id, status, pid, bundle, createdAt }`.
    
    - `status` ∈ { `"creating"`, `"running"`, `"stopped"` }.
        
- **Exit codes**: MRUNC exits with the child’s exit code; runtime-level failures return non-zero codes (documented in MRUNC docs).
    

---

## 3) Behavior that deviates (by design)

- **No hooks** (`prestart`, `poststart`, `poststop`) — not implemented.
    
- **No network namespace** — out of scope entirely (no veth/bridge/ip rules).
    
- **No IPC/time/cgroup namespaces** — not created.
    
- **No AppArmor/SELinux labels** — not applied.
    
- **No “update” API** (live cgroup updates) — out of scope.
    
- **No additional `exec`** into running container — out of scope.
    
- **No `setns`** into existing namespaces (`namespaces[].path` ignored).
    

---

## 4) Filesystem policy (pivot_root & mounts)

- MRUNC always:
    
    1. `unshare(CLONE_NEWNS)` then set `/` propagation to **`MS_PRIVATE`** (recursive).
        
    2. Mount new root (from `root.path`) with **`ro,nodev,nosuid`** where possible.
        
    3. `pivot_root(newroot, newroot/put_old)`; `chdir("/")`; `umount2("/put_old", MNT_DETACH)`; remove `/put_old`.
        
    4. Mount `proc` **after** PID ns: `proc` at `/proc` with `nosuid,noexec,nodev,hidepid=2`.
        
    5. Mount a **minimal `/dev`** on tmpfs + `devpts` for TTY if `terminal=true`.
        
    6. For any configured bind mounts: enforce `nosuid,nodev,noexec` unless explicitly overridden by a “dangerous” flag (not recommended).
        
    7. Default **rootfs read-only**; MRUNC will allow `rw` only for whitelisted paths (`/tmp`, `/run` via tmpfs).
        
- **`chroot` is never used**; MRUNC uses `pivot_root` explicitly.
    

---

## 5) Security policy (caps + seccomp)

- **Capabilities:**
    
    - Default: **drop all** (empty effective/bounding/inheritable).
        
    - If `process.capabilities` provided, MRUNC grants only those listed in `bounding` + `effective` (intersection with a hard denylist that always excludes highly dangerous caps like `CAP_SYS_ADMIN` unless MRUNC is run in a special “unsafe” mode).
        
- **Seccomp:**
    
    - If a profile is provided: apply it exactly (only allow `SCMP_ACT_ALLOW` sets, default deny).
        
    - If not provided: MRUNC loads its **default allowlist** sufficient for BusyBox/shell and most non-privileged CLI tools; explicitly **denies** `mount`, `umount2`, `pivot_root`, `ptrace`, `kexec_*`, `bpf`, `perf_event_open`, `keyctl`, `userfaultfd`, `open_by_handle_at`, `reboot`, `unshare` (re-namespacing), module load, etc.
        
- **noNewPrivileges:** always **true**.
    

---

## 6) Minimal error model

- Invalid `type` in `mounts[]` → error: `unsupported mount type`.
    
- Using unsupported namespace types (`network`, `ipc`, `time`, `cgroup`) → error.
    
- `linux.resources` cgroup v1 fields present → ignored with warning (MRUNC uses v2 only).
    
- Missing `root.path` or missing `process.args` → error.
    
- Capabilities requested but blocked by MRUNC hard denylist → error.
    

---

## 7) Conformance & tests (what MRUNC will run)

- **Bundle validation:** presence of `config.json`, `rootfs/`.
    
- **Happy-path**: run `/bin/true`, `/bin/sh -lc 'echo hi'`.
    
- **Namespaces**: inside container `ps` shows PID 1; `hostname` equals `config.json.hostname`; user IDs per mapping.
    
- **FS isolation**: host paths not visible; `/proc` mounted with `hidepid=2`.
    
- **Caps/Seccomp**: attempts to `mount`, `ptrace`, `bpf`, `userfaultfd` → fail (EPERM/KILL).
    
- **Cgroups v2**: `pids.max` enforced; `memory.max` enforced; `cpu.max` throttles.
    

---

## 8) Example minimal `config.json` (MRUNC-subset compliant)

```json
{
  "ociVersion": "1.1.0",
  "hostname": "mrunc-demo",
  "root": { "path": "rootfs", "readonly": true },
  "process": {
    "terminal": false,
    "cwd": "/",
    "args": ["/bin/sh", "-lc", "echo hello from mrunc && sleep 1"],
    "env": ["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],
    "user": { "uid": 0, "gid": 0 },
    "noNewPrivileges": true,
    "capabilities": {
      "bounding": [], "effective": [], "inheritable": []
    }
  },
  "mounts": [
    { "destination": "/proc", "type": "proc", "source": "proc", "options": ["nosuid","noexec","nodev"] },
    { "destination": "/tmp",  "type": "tmpfs", "source": "tmpfs", "options": ["nosuid","nodev","mode=1777"] }
  ],
  "linux": {
    "uidMappings": [{"hostID": 1000, "containerID": 0, "size": 1}],
    "gidMappings": [{"hostID": 1000, "containerID": 0, "size": 1}],
    "namespaces": [{"type": "mount"},{"type": "pid"},{"type": "uts"},{"type": "user"}],
    "resources": {
      "pids": { "limit": 64 },
      "cpu": { "max": "50000 100000" },
      "memory": { "limit": 536870912 }
    },
    "seccomp": {
      "defaultAction": "SCMP_ACT_KILL",
      "syscalls": [
        {
          "names": ["read","write","exit","futex","nanosleep","clock_gettime","rt_sigaction","rt_sigprocmask","brk","mmap","mprotect","munmap","close","dup","dup2","fcntl","ioctl","getpid","getppid","getuid","getgid","set_tid_address","set_robust_list","prlimit64","execve"],
          "action": "SCMP_ACT_ALLOW"
        }
      ]
    }
  }
}
```

> Note: For the `uid/gid` mappings example above, MRUNC expects `/proc/sys/kernel/unprivileged_userns_clone` to allow userns, and that the host has uid/gid `1000` (typical). Adjust as needed.

---

## 9) One-page “subset matrix” (for your report)

|Spec area|MRUNC status|Notes|
|---|---|---|
|Bundle layout|✅|Standard `config.json` + `rootfs/`|
|Lifecycle (`create/start/run/delete`)|✅|`state` optional|
|`process`|✅ (subset)|args/cwd/env/user/caps/noNewPriv|
|`root`|✅|`readonly` default true|
|`mounts`|✅ (subset)|`bind`, `proc`, `tmpfs`, `devpts` only|
|Namespaces|✅ (subset)|`mount`, `pid`, `uts`, `user` only|
|Cgroups|✅ (v2 only)|pids, cpu, memory, (simple io if provided)|
|Seccomp|✅ (allowlist)|Default deny; allow minimal set|
|Capabilities|✅ (subset)|Start empty; add only requested & safe|
|Hooks|❌|Not implemented|
|Networking|❌|Out of scope|
|LSM (AppArmor/SELinux)|❌|Ignored|
|Update/exec API|❌|Not supported|
|Windows/Solaris|❌|Linux-only|

---

If you want, I can also generate:

- a **starter default seccomp.json** (allowlist tuned for BusyBox/bash),
    
- and a **YAML test plan** that turns the matrix above into runnable checks.