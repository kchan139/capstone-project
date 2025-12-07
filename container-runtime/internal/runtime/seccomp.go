package runtime

import (
    "fmt"
    seccomp "github.com/seccomp/libseccomp-golang"
    mySpecs "mrunc/pkg/specs"
)

func SetupSeccomp(config *mySpecs.SeccompConfig) error {
    if config == nil {
        return nil // No seccomp profile specified
    }

    // Parse default action
    defaultAction, err := parseAction(config.DefaultAction)
    if err != nil {
        return fmt.Errorf("invalid default action: %v", err)
    }

    // Create filter with default action
    filter, err := seccomp.NewFilter(defaultAction)
    if err != nil {
        return fmt.Errorf("failed to create seccomp filter: %v", err)
    }
    defer filter.Release()

	// Add architectures if specified
    if len(config.Architectures) > 0 {
        // Remove default native arch first
        if err := filter.RemoveArch(seccomp.ArchNative); err != nil {
            return fmt.Errorf("failed to remove native arch: %v", err)
        }

        for _, archStr := range config.Architectures {
            arch, err := parseArch(archStr)
            if err != nil {
                return fmt.Errorf("invalid architecture %s: %v", archStr, err)
            }
            if err := filter.AddArch(arch); err != nil {
                return fmt.Errorf("failed to add arch %s: %v", archStr, err)
            }
        }
    }

    // Add syscall rules
    for _, call := range config.Syscalls {
        action, err := parseAction(call.Action)
        if err != nil {
            return fmt.Errorf("invalid action %s: %v", call.Action, err)
        }

        for _, name := range call.Names {
            syscallID, err := seccomp.GetSyscallFromName(name)
            if err != nil {
                // Syscall might not exist on this architecture
                continue
            }

            if err := filter.AddRule(syscallID, action); err != nil {
                return fmt.Errorf("failed to add rule for %s: %v", name, err)
            }
        }
    }

    // Load the filter into the kernel
    if err := filter.Load(); err != nil {
        return fmt.Errorf("failed to load seccomp filter: %v", err)
    }

    return nil
}

// parseAction converts OCI action string to seccomp action
func parseAction(action string) (seccomp.ScmpAction, error) {
    switch action {
    case "SCMP_ACT_KILL":
        return seccomp.ActKill, nil
    case "SCMP_ACT_KILL_PROCESS":
        return seccomp.ActKillProcess, nil
    case "SCMP_ACT_KILL_THREAD":
        return seccomp.ActKillThread, nil
    case "SCMP_ACT_TRAP":
        return seccomp.ActTrap, nil
    case "SCMP_ACT_ERRNO":
        return seccomp.ActErrno, nil
    case "SCMP_ACT_TRACE":
        return seccomp.ActTrace, nil
    case "SCMP_ACT_ALLOW":
        return seccomp.ActAllow, nil
    case "SCMP_ACT_LOG":
        return seccomp.ActLog, nil
    default:
        return seccomp.ActKill, fmt.Errorf("unknown action: %s", action)
    }
}

func parseArch(arch string) (seccomp.ScmpArch, error) {
    switch arch {
    case "SCMP_ARCH_X86_64":
        return seccomp.ArchAMD64, nil
    case "SCMP_ARCH_X86":
        return seccomp.ArchX86, nil
    case "SCMP_ARCH_AARCH64":
        return seccomp.ArchARM64, nil
    case "SCMP_ARCH_ARM":
        return seccomp.ArchARM, nil
    case "SCMP_ARCH_MIPS":
        return seccomp.ArchMIPS, nil
    case "SCMP_ARCH_MIPS64":
        return seccomp.ArchMIPS64, nil
    case "SCMP_ARCH_MIPS64N32":
        return seccomp.ArchMIPS64N32, nil
    case "SCMP_ARCH_MIPSEL":
        return seccomp.ArchMIPSEL, nil
    case "SCMP_ARCH_MIPSEL64":
        return seccomp.ArchMIPSEL64, nil
    case "SCMP_ARCH_MIPSEL64N32":
        return seccomp.ArchMIPSEL64N32, nil
    case "SCMP_ARCH_PPC":
        return seccomp.ArchPPC, nil
    case "SCMP_ARCH_PPC64":
        return seccomp.ArchPPC64, nil
    case "SCMP_ARCH_PPC64LE":
        return seccomp.ArchPPC64LE, nil
    case "SCMP_ARCH_S390":
        return seccomp.ArchS390, nil
    case "SCMP_ARCH_S390X":
        return seccomp.ArchS390X, nil
    case "SCMP_ARCH_RISCV64":
        return seccomp.ArchRISCV64, nil
    default:
        return seccomp.ArchInvalid, fmt.Errorf("unknown architecture: %s", arch)
    }
}