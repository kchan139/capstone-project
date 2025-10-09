## Overview: What Layer We’re Comparing

|Layer|Example|Description|
|---|---|---|
|**Container Engine**|Docker, Podman|Provides image management, networking, registry, and orchestration features.|
|**Container Runtime**|**runc**, **crun**, **youki**, **MRUNC**|Implements the **OCI Runtime Specification**: creating namespaces, setting cgroups, pivot_root, applying seccomp/capabilities, and starting the process.|

So MRUNC, runc, and crun are all in the **same layer** — they actually perform similar steps — but differ in _scope_, _language_, and _design philosophy_.

---

## MRUNC vs runc vs crun — Minimalism Comparison

|Category|**runc** (Go, OCI reference)|**crun** (C, OCI runtime)|**MRUNC** (Go, educational minimalist)|
|---|---|---|---|
|**Purpose**|Production-grade, fully OCI-compliant runtime used by Docker, Podman, containerd|Faster, lightweight alternative to runc for production|Minimalist, educational runtime to explore Linux isolation & sandboxing|
|**OCI Compliance**|✅ Full (passes OCI tests, supports lifecycle hooks, annotations, networking, mounts, etc.)|✅ Full (same features, optimized)|⚠️ Partial (only implements core process & linux sections; skips hooks, networking, and annotations)|
|**Supported Namespaces**|All 7 (mount, PID, UTS, IPC, network, user, cgroup)|All 7|Only 4 core ones (mount, PID, UTS, user)|
|**Cgroups Support**|v1 & v2|v1 & v2 (fast libocispec parsing)|v2 only (simpler, cleaner unified hierarchy)|
|**Security Features**|Seccomp, capability management, AppArmor/SELinux integration, rootless mode, no-new-privs, ambient caps|Same as runc, with slightly faster parsing|Only seccomp + capabilities + no-new-privs (no AppArmor/SELinux/rootless)|
|**Networking**|Full network namespace handling (veth, bridge setup, etc., via libnetwork or upper layer)|Same|❌ No networking (out of scope)|
|**Hooks / Lifecycle**|Prestart, poststart, poststop hooks|Same|❌ Not implemented (simplicity)|
|**Language & Dependencies**|Go + large dependency chain (runtime-spec, libcontainer, etc.)|C + libocispec, libcap, libseccomp|Go only, direct `syscall`/`unix` usage (no external deps)|
|**Performance**|Moderate startup (~40–70 ms typical)|Faster startup (~20–30 ms) due to C implementation|Expected to be **fastest** (no JSON parsing layers, no daemon), but not yet benchmarked|
|**Complexity / Codebase Size**|~25k+ LOC (libcontainer + CLI)|~20k LOC|~2–3k LOC (planned)|
|**Goal of Minimalism**|Feature completeness|Efficiency and performance|**Simplicity, auditability, and teachability**|

---

## Analysis

### ** Functional Minimalism**

- **runc/crun** implement _everything_ from the OCI spec — including hooks, annotations, mount options, networking, rootless support, and JSON parsing via full OCI schema.
    
- **MRUNC** strips away these layers to keep **only what proves isolation and security**.  
    So MRUNC’s minimalism is **scope-based**, not functionality-broken — it runs real containers, just not all OCI features.
    

---


    



### ** Security Minimalism**

- **runc/crun** support AppArmor, SELinux, seccomp, capability bounding, rootless mode, and no-new-privs.  
    ➤ They maximize _coverage_, not simplicity.
    
- **MRUNC** intentionally **minimizes the attack surface** by _removing_ complex subsystems like networking and IPC, which are the most common container escape vectors.  
    ➤ Its seccomp + cap-drop model covers the essentials without the overhead of LSMs.
    

---

### ** Performance Minimalism**

- **crun** already outperforms **runc** thanks to C’s lower overhead and faster OCI config parsing.
    
- **MRUNC**, by removing JSON validation, hook management, and daemon coordination, may be even lighter for _single-container educational runs_.  
    ➤ It doesn’t beat crun in scalability, but in **startup latency and footprint per container**, it could be near zero overhead.
    

---

### ** Pedagogical Minimalism**

- **runc/crun** are unreadable for most students (tens of thousands of lines, complex flow).
    
- **MRUNC** is structured for **clarity over completeness**, allowing students to trace the life of a container creation:
    
    1. Parse JSON config →
        
    2. Fork →
        
    3. Apply namespaces →
        
    4. pivot_root →
        
    5. Drop caps/seccomp →
        
    6. Execute process.
        

This makes MRUNC not just smaller, but _educationally transparent_.

---

## Summary

|Aspect|MRUNC|runc|crun|
|---|---|---|---|
|**Design goal**|Educational, secure-by-simplicity|Reference implementation|Fast, efficient production runtime|
|**LOC (approx.)**|~3K|~25K|~20K|
|**OCI compliance**|Partial|Full|Full|
|**Namespaces**|4|7|7|
|**Cgroups**|v2 only|v1 + v2|v1 + v2|
|**Security scope**|seccomp + caps only|seccomp + caps + LSMs + rootless|same as runc|
|**Networking**|❌ none|✅ full|✅ full|
|**Goal of minimalism**|simplicity, security, pedagogy|correctness, completeness|performance, compatibility|

---

###  In short**

> MRUNC is “minimal” not because it lacks features —  
> but because it focuses on **only the primitives** necessary to **demonstrate and study** isolation,  
> whereas **runc** and **crun** aim to **run production workloads**.
