# Web panel

```bash
wireproxy -c /etc/wireproxy/config --admin 127.0.0.1:9090 --admin-token "$(openssl rand -hex 24)"
```

Open `http://127.0.0.1:9090` and sign in with the token (if you omit `--admin-token`, a random one is generated and
printed to the log on start). The panel is a single embedded page, no extra files or internet access needed.

- **Status** - tunnel state, uptime, every peer's endpoint, last handshake and traffic, and the active listeners.
- **Config** - edit the config file in the browser. `PrivateKey`, `PresharedKey`, `HeaderProtectionKey` and
  proxy `Password` values are shown as `********` (tick *Show secrets* to see them); leaving the placeholder keeps the
  stored value. *Validate* parses the text exactly like the daemon does, *Save* writes it atomically and keeps the
  previous file as `<config>.bak`, *Save & apply* also restarts the daemon. There are helpers to insert an AmneziaWG
  parameter template and to generate a `HeaderProtectionKey`.
- **Logs** - live tail of the daemon log (empty with `--silent`).

Applying a config restarts the process in place (about a second of downtime; the same PID on Linux/macOS).
`kill -HUP <pid>` does the same from a shell. If the new config does not start, the panel stays up, shows the error
and offers *Restore previous config*.

The same API is available for scripting with `Authorization: Bearer <token>`:
`GET /api/status`, `GET|PUT /api/config`, `POST /api/validate`, `POST /api/reload`, `GET /api/logs`,
and `GET /healthz` (no auth, for health checks).

**Security notes.** The panel can read and replace your private keys, so treat the token like a password.
It speaks plain HTTP: keep it on `127.0.0.1` or put it behind a TLS reverse proxy (the daemon warns when the
address is reachable from the network). Login attempts are rate limited, sessions use `HttpOnly`/`SameSite=Strict`
cookies and a strict Content-Security-Policy. Because the process must write the config and restart itself, `--admin`
turns off the landlock/pledge sandbox that plain mode applies. `--admin` cannot be combined with `--daemon`.
In Docker the config directory must be writable by the container user (`sudo chown -R 65532:65532 config`).
