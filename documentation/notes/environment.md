# Infrastructure-as-Code (IaC) Rationale

## Purpose of `environment/`

This project involves **low-level container runtime programming** that directly manipulates Linux namespaces, mounts, and root filesystems.  
Such experiments **can** break or corrupt the operating system if misconfigured — especially when run as `root`.

The `environment/` directory exists to provide an **automated, repeatable, and secure way** to rebuild the development environment from scratch.

---

## Why Terraform + Ansible?

1. **Instant Recovery from OS Breakage**  
   - If a droplet is rendered unusable during testing, it can be destroyed and recreated in minutes.
   - No manual setup steps — all provisioning is automated.

2. **Consistent Baseline**  
   - Every rebuild includes the same security hardening, firewall rules, users, and tooling.
   - Ensures reproducibility for both development and demonstration.

3. **Security by Default**  
   - SSH key-based login only, custom ports, root login disabled, firewall locked down.
   - Prevents misconfigurations from leaving the system exposed after a rebuild.

4. **Disposable Test Environments**  
   - Multiple identical droplets can be created for destructive tests without risking the main development server.

---

## What’s Inside

- **`terraform/`** → Creates a DigitalOcean droplet and configures networking + firewall rules.
- **`ansible/`** → Provisions the droplet with necessary tools (Docker, VSCode Server, etc.) and applies security hardening.
- **`vsc-server/`** → Docker-based remote development environment.

---

## TL;DR
`environment/` is our **reset button**.  
If (when) something breaks during low-level OS experiments, we can rebuild a secure, ready-to-use development server in minutes.
