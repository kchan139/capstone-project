# Lightweight sandbox for malware/risky code”

### 1  threat model 

- **Assets to protect:** host kernel, host filesystem, host network, other containers.
    
- **Adversary:** untrusted code inside the container that tries to escape, abuse kernel bugs, or exfiltrate data.
    
- **Attacker capabilities:** arbitrary code execution as root _inside_ the container; attempts at namespace reconfiguration, mount tricks, and dangerous syscalls.
    
- **Out of scope:** hypervisor-grade isolation (that’s Kata’s domain), kernel 0-day with no mitigation, and LSM policy authoring (AppArmor/SELinux) if kept out of scope.
    

### 2 Security objectives (what MRUNC provides)

- **Process isolation:** PID/UTS/User/Mount namespaces; no visibility of host PIDs; mapped root via userns.
    
- **FS isolation:** `pivot_root` with a **private** mount tree; mount options: `nodev,nosuid,noexec,ro` where possible; read-only bind of `/usr`, tmpfs for `/tmp`, carefully mounted `/proc` and `/sys` (only what’s needed).
    
- **Capability minimization:** start from **empty** capability set, then add only what the entry process needs (often none). Avoid `CAP_SYS_ADMIN` at all costs.
    
- **Syscall reduction:** default **seccomp-BPF** profile (deny by default, allowlist) aligned with industry practice; remove high-risk syscalls (e.g., `mount`, `ptrace`, `kexec_load`, raw `bpf`/`perf_event_open`, `unshare`), and restrict `clone3`/`userfaultfd` unless required. (Compare with how ChromeOS Minijail/bwrap/nsjail do it.) [GitHub+3chromium.googlesource.com+3Google GitHub+3](https://chromium.googlesource.com/chromiumos/docs/%2B/master/sandboxing.md?utm_source=chatgpt.com)
    
- **Resource containment:** cgroups v2 limits (CPU/mem/IO) to prevent DoS.
    
- **No networking (by design):** biggest attack channel removed; makes the sandbox safer by default.
    

### 3 Design decisions that “lock down” escape paths

- Make the container’s mount namespace **MS_PRIVATE** before any `pivot_root`.
    
- Mount `/proc` with **`hidepid=2,nosuid,noexec,nodev`** (and only after PID ns is active).
    
- Bind-mount a tiny `/dev` with just `null`, `zero`, `random`, `urandom`, `tty` as needed; all **`nodev`** where possible.
    
- **Drop setuid binaries** in the rootfs build step.
    
- Optional “hardening mode” flag: make rootfs **read-only** and remount per-path `rw` exceptions.
    
- Provide a **modeled default seccomp policy** and a tooling script to auto-diff your policy vs. gVisor/minijail/bwrap typical allows to justify choices (cite comparable projects). [GitHub+3gvisor.dev+3gvisor.dev+3](https://gvisor.dev/docs/?utm_source=chatgpt.com)
    

### 4 Benchmark (methods & metrics)

- **Functional isolation tests:** “escape attempts” corpus (mount, re-unshare, ptrace, /proc poking, keyctl, etc.). Expect **EPERM/EACCES/KILL**.
    
- **Red team scripts:** run known container-escape PoC _patterns_ (sanitized) to verify they are neutered by policy (no real 0-days).
    
- **Coverage diff:** run target under `strace` to confirm only allowed syscalls appear.
    
- **Performance overhead:** measure `time-to-exec /bin/true`, RSS of runtime process, and 60-sec CPU burn with/without seccomp/caps to show “lightweight”.