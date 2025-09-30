# OCI Runtime Specification Compliance Map

**Last Updated:** [Sep 29, 2025]
**MRUNC Version:** [0.1.1]

---

## Executive Summary

### Compliance Status
- **Overall Compliance:** [X]% implemented
- **Target Compliance Level:** [Full/Partial/Inspired-by]
- **Primary Goal:** [Educational/Production/Research]

### Quick Stats
- ✅ Implemented: [X] features
- ⚠️ Partially Implemented: [X] features
- ❌ Not Implemented: [X] features
- 🔮 Planned: [X] features

---

## 1. Configuration Structure

### 1.1 Container Configuration (config.json)

| OCI Requirement | Status | MRUNC Implementation | Notes |
|----------------|--------|---------------------|-------|
| `ociVersion` | ❌ | Not present | We use custom format |
| `root.path` | ✅ | `root.path` | Implemented |
| `root.readonly` | ✅ | `root.readonly` | Implemented |
| `process.args` | ✅ | `process.args` | Implemented |
| `process.env` | ✅ | `process.env` | Implemented |
| `process.cwd` | ✅ | `process.cwd` | Implemented |
| `process.terminal` | ✅ | `process.terminal` | Implemented |
| `process.user` | ✅ | `process.user` | Implemented (uid, gid) |
| `process.capabilities` | ❌ | Not implemented | Planned |
| `process.rlimits` | ❌ | Not implemented | |
| `process.noNewPrivileges` | ❌ | Not implemented | |
| `process.apparmorProfile` | ❌ | Not implemented | |
| `process.selinuxLabel` | ❌ | Not implemented | |
| `hostname` | ✅ | `hostname` | Implemented |
| `mounts` | ❌ | Hardcoded (proc only) | |
| `hooks` | ❌ | Not implemented | |
| `annotations` | ❌ | Not implemented | |
| `linux` | ⚠️ | Partially | See section 2 |

**Deviation Rationale:**
- [Explain why config format differs]
- [Document conscious design choices]

---

## 2. Linux-Specific Configuration

### 2.1 Namespaces

| Namespace Type | OCI Spec | Status | Implementation Location |
|---------------|----------|--------|------------------------|
| PID | Required | ✅ | `runtime/namespace.go:12` |
| Network | Required | ❌ | Not implemented |
| Mount | Required | ✅ | `runtime/namespace.go:12` |
| IPC | Required | ❌ | Not implemented |
| UTS | Required | ✅ | `runtime/namespace.go:12` |
| User | Optional | ❌ | Attempted but incomplete |
| Cgroup | Optional | ❌ | Not implemented |

**Implementation Notes:**
```go
// Current implementation
Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS
```

**Issues:**
- [Document any namespace-related bugs or limitations]

### 2.2 Control Groups (cgroups)

| Feature | Status | Notes |
|---------|--------|-------|
| cgroups v2 support | ❌ | Mentioned in README but not implemented |
| Resource limits (CPU) | ❌ | |
| Resource limits (Memory) | ❌ | |
| Resource limits (I/O) | ❌ | |
| Device access control | ❌ | |

### 2.3 Security

| Feature | OCI Spec | Status | Notes |
|---------|----------|--------|-------|
| Seccomp profiles | Optional | ❌ | Mentioned in README |
| Capabilities | Optional | ❌ | |
| AppArmor | Optional | ❌ | |
| SELinux | Optional | ❌ | |
| No new privileges | Optional | ❌ | |
| Masked paths | Optional | ❌ | |
| Readonly paths | Optional | ❌ | |

### 2.4 User and Group Management

| Feature | Status | Implementation | Notes |
|---------|--------|---------------|-------|
| User ID mapping | ✅ | `runtime/user.go` | Basic implementation |
| Group ID mapping | ✅ | `runtime/user.go` | Basic implementation |
| Additional GIDs | ⚠️ | Supported in config | Not tested |
| UID/GID validation | ✅ | `runtime/user.go:ValidateUser()` | Range: 0-65535 |

---

## 3. Filesystem

### 3.1 Root Filesystem

| Requirement | Status | Implementation | Notes |
|------------|--------|---------------|-------|
| pivot_root | ✅ | `runtime/filesystem.go` | Using raw syscall |
| Bind mount support | ⚠️ | Hardcoded for rootfs | |
| Mount propagation | ❌ | Not implemented | |
| rootfs readonly option | ✅ | Config supported | Not enforced |

**Current Implementation:**
```go
// From child.go
unix.Mount(root_fs, root_fs, "", unix.MS_BIND, "")
PivotRoot(root_fs, root_fs_putold)
```

### 3.2 Mounts

| Feature | OCI Spec | Status | Notes |
|---------|----------|--------|-------|
| Procfs | Standard | ✅ | Hardcoded in child.go |
| Sysfs | Standard | ❌ | |
| Devfs | Standard | ❌ | |
| Tmpfs | Standard | ❌ | |
| Custom mounts | Required | ❌ | |
| Mount options | Required | ❌ | |

---

## 4. Runtime and Lifecycle

### 4.1 State Machine

OCI defines container states: creating → created → running → stopped

| State | Status | Notes |
|-------|--------|-------|
| Creating | ⚠️ | Implicit, not exposed |
| Created | ❌ | No separate create command |
| Running | ✅ | Via `run` command |
| Stopped | ⚠️ | Implicit when process exits |

**Current Lifecycle:**
```
run command → child process → exec → [container runs] → exit
```

### 4.2 Runtime Operations

| Operation | OCI Spec | Status | Implementation |
|-----------|----------|--------|---------------|
| create | Required | ❌ | Combined with start |
| start | Required | ⚠️ | Merged into `run` |
| kill | Required | ❌ | |
| delete | Required | ❌ | |
| state | Required | ❌ | |

**Deviation Notes:**
- We use a single `run` command instead of separate create/start
- No container ID tracking
- No state persistence

---

## 5. Process Execution

### 5.1 Command Execution

| Feature | Status | Implementation | Notes |
|---------|--------|---------------|-------|
| Exec with PATH resolution | ✅ | `runtime/ExecuteCommand.go` | Fallback chain |
| Absolute path execution | ✅ | `runtime/ExecuteCommand.go:tryDirectExec()` | |
| Environment variables | ✅ | `cli/child.go` | Full support |
| Working directory | ✅ | `cli/child.go` | Via chdir |
| Shell fallback | ⚠️ | Implemented | Security warning logged |

### 5.2 Standard Streams

| Feature | Status | Notes |
|---------|--------|-------|
| stdin (interactive) | ✅ | When terminal=true |
| stdout | ✅ | Connected |
| stderr | ✅ | Connected |
| Non-interactive mode | ✅ | Stdin detached when terminal=false |

---

## 6. Inter-Process Communication

### 6.1 Parent-Child Communication

| Mechanism | Purpose | Implementation |
|-----------|---------|---------------|
| Pipe (config transfer) | Send config to child | `cli/run.go`, `cli/child.go` |
| Environment variable | Pass pipe FD | `_MRUNC_PIPE_FD` |
| Process wait | Lifecycle sync | `cmd.Wait()` |

**Configuration Passing:**
```go
// Parent serializes config → pipe → Child deserializes
configData, _ := json.Marshal(config)
parentPipe.Write(configData)
```

---

## 7. Known Gaps and Limitations

### Critical Missing Features
1. **No container isolation tracking** - No way to list/manage running containers
2. **Missing network namespace** - Containers share host network
3. **No cgroups** - No resource limits
4. **No IPC namespace** - Shared IPC with host
5. **No hooks support** - Can't inject custom logic at lifecycle events

### Security Concerns
1. No seccomp filtering
2. No capability dropping
3. Shell fallback in command execution
4. Running as root required

### Functional Limitations
1. Single rootfs per config (no layers/overlays)
2. No mount management
3. No device management
4. Config format incompatible with OCI bundles

---

## 8. Future Roadmap

### High Priority (Must Have)
- [ ] Implement network namespace
- [ ] Add IPC namespace
- [ ] Basic cgroups v2 support (CPU, memory limits)
- [ ] Eliminate shell fallback

### Medium Priority (Should Have)
- [ ] Seccomp profiles
- [ ] Capability dropping
- [ ] Standard mount points (dev, sys)
- [ ] Container state tracking

### Low Priority (Nice to Have)
- [ ] Full OCI config.json compatibility
- [ ] Hooks support
- [ ] User namespace improvements
- [ ] AppArmor/SELinux support

### Out of Scope
- [ ] Image management
- [ ] Registry support
- [ ] Networking plugins
- [ ] Volume management

---

## 9. Testing Against OCI

### Test Results
- [ ] OCI runtime tools validation: [Not attempted/Pass/Fail]
- [ ] runc comparison test: [Status]
- [ ] OCI bundle compatibility: [Status]

### Test Commands
```bash
# Add commands used to test OCI compliance
```

---

## 10. References

- [OCI Runtime Specification](https://github.com/opencontainers/runtime-spec/blob/main/spec.md)
- [OCI Runtime Config](https://github.com/opencontainers/runtime-spec/blob/main/config.md)
- [OCI Runtime Linux Spec](https://github.com/opencontainers/runtime-spec/blob/main/config-linux.md)

---

## Notes

**Purpose of This Document:**
This document maps our implementation against the OCI Runtime Specification to:
1. Track what we've built vs. what the standard defines
2. Document conscious deviations and their rationale
3. Guide future development priorities
4. Serve as reference for thesis defense/documentation

**How to Use:**
- Update this as features are added
- Mark implementation locations for easy reference
- Document *why* we deviate, not just *that* we deviate
- Keep the "Known Gaps" section honest for academic integrity
