# ssm

SSH connection manager. Save your hosts, pick one, connect.

## Install

```
curl -fsSL https://gossm.sh/install.sh | sh
```

Or build from source:

```
go install ./cmd/ssm
```

## Usage

```
ssm
ssm add
ssm edit <name>
ssm remove <name>
```

Arrow keys to navigate, `/` to search, `enter` to connect.

## How it works

Connections are stored in `~/.config/ssm/connections.enc`, encrypted with AES-256-GCM + Argon2id. A master password is required to unlock.

SSH sessions use a native Go client with PTY and window resize support. No `sshpass` or external dependencies.

## Cloud sync (optional)

Disabled by default. If you want to sync across machines:

```
ssm register
ssm login
ssm push
ssm pull
```

The server never sees your data. You can also self-host it.

## License

MIT
