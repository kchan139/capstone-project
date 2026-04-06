# `environment/` overview

`environment/` contains infrastructure, provisioning, and remote development files.

## Terraform

`environment/terraform/` defines:

- the DigitalOcean provider,
- one droplet named `capstone-project`,
- image `ubuntu-24-04-x64`,
- region `sgp1`,
- size `s-2vcpu-4gb`,
- two SSH keys attached to the droplet,
- one firewall attached to that droplet.

The firewall currently allows:

- inbound TCP on `var.custom_ssh_port`,
- inbound ICMP,
- outbound TCP, UDP, and ICMP.

## Ansible

`environment/ansible/playbook.yml` provisions the server with:

- stopped and masked unattended update services,
- base packages such as `make`, `curl`, `tree`, `bat`, and `software-properties-common`,
- timezone `Asia/Ho_Chi_Minh`,
- two sudo users from vault variables,
- UFW rules,
- SSH on port `22` and the custom SSH port,
- disabled password authentication and root login,
- Starship, Helix, Yazi, and Go,
- Go workspace directories for both users,
- `gopls`, `dlv`, and `goimports` in `/usr/local/bin`,
- `/var/lib/mrunc/images/`,
- an Ubuntu 24.04 minimal rootfs extracted into `/var/lib/mrunc/images/ubuntu`.

## Helper scripts

`environment/ansible/scripts/` contains:

- `run.sh` — builds an inventory from Terraform output and runs the playbook as `root`.
- `re-run.sh` — rebuilds inventory and reruns the playbook with become.
- `generate-ssh-keys.sh` — helper referenced by the key generation playbook.

These scripts expect local access to `tofu`, `jq`, `ansible`, and `ansible-vault`.

## OpenVSCode

`environment/vsc-server/compose.yml.j2` contains the OpenVSCode Server compose template used by the remote development setup.
