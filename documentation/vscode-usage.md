# VS Code Server Setup Guide

This document explains how to use the OpenVSCode Server (`vsc-server`) provisioned via Ansible in this project.

## Prerequisites

- SSH access to the server.

## Accessing VS Code Server Container

### 1. SSH into the server:
```bash
ssh -L <local_port>:127.0.0.1:<remote_port> <username>@<server_ip> -p <ssh_port>
```
#### Notes

- `<local_port>` → any free port on your local machine; the port you will use in your browser.
- `<remote_port>` → the port that the VS Code container is listening on the host (as defined in `compose.yml`).
- `<username>` → server username
- `<server_ip>` → server’s public IP
- `<ssh_port>` → custom SSH port on the server

### 2. Each user has a dedicated Docker Compose setup located in their home directory:
```
/home/<username>/compose.yml
```

### 3. Start the OpenVSCode Server container:

```bash
docker compose up -d
```

### 4. Access VS Code in your browser at:
```
http://localhost:<local_port>
```

## Stopping the Container

To stop VS Code for your user:

```bash
docker compose down
```

## Updating VS Code Server

If you want to update VS Code:

```bash
docker compose pull
docker compose up -d
```

## Notes

* All workspace files are mounted to `~/workspace` in the container.

---

> This setup ensures isolated VS Code environments per user with Docker and proper UID/GID mapping.
