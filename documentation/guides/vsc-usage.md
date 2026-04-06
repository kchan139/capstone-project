# SSH forwarding and OpenVSCode

This repo includes:

- `port-forward.sh` for SSH local port forwarding.
- `environment/vsc-server/compose.yml.j2` as an OpenVSCode Server compose template.

## `.env` file

`port-forward.sh` reads connection details from `.env` in the repo root.

Start from the template:

```bash
cp .env.example .env
```

Set:

- `USERNAME`
- `SERVER_IP`
- `SSH_PORT`
- `PORTS`
- `REMOTE_HOST`

`PORTS` is a space-separated list of `<local-port>:<remote-port>` mappings.

Example:

```bash
PORTS="8080:80 3000:3000"
REMOTE_HOST="127.0.0.1"
```

## Start port forwarding

```bash
./port-forward.sh
```

The script expands each entry in `PORTS` into an SSH `-L` rule and opens an SSH session to `$USERNAME@$SERVER_IP` on `$SSH_PORT`.

## OpenVSCode compose template

`environment/vsc-server/compose.yml.j2` defines one service:

- service name: `openvscode`
- image: `gitpod/openvscode-server`
- container port: `3000`
- host bind: `127.0.0.1:{{ item.port }}:3000`
- workspace mount: `.:/home/workspace:cached`
- runtime user: `{{ item.uid }}:{{ item.gid }}`
- `stdin_open: true`
- `tty: true`
- `restart: unless-stopped`
