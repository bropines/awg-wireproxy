# awg-wireproxy

[![Test](https://github.com/bropines/awg-wireproxy/actions/workflows/test.yml/badge.svg)](https://github.com/bropines/awg-wireproxy/actions/workflows/test.yml)
[![Release](https://img.shields.io/github/v/release/bropines/awg-wireproxy?include_prereleases)](https://github.com/bropines/awg-wireproxy/releases)
[![Docker](https://img.shields.io/badge/ghcr.io-awg--wireproxy-blue?logo=docker)](https://github.com/bropines/awg-wireproxy/pkgs/container/awg-wireproxy)
[![ISC licensed](https://img.shields.io/badge/license-ISC-blue)](./LICENSE)

A WireGuard **and AmneziaWG** client that exposes itself as a socks5/http proxy or tunnels - completely in
userspace, no root, no network interface.

## What is this

`wireproxy` connects to a WireGuard/AmneziaWG peer and exposes a socks5/http proxy or tunnels on the machine.
It is useful if you need to reach certain sites through a WireGuard peer, but cannot or do not want to set up a
new network interface. The core is [amneziawg-go](https://github.com/amnezia-vpn/amneziawg-go) **v3.1**, so
plain WireGuard configs work unchanged and every AmneziaWG obfuscation option (1.0, 1.5, 2.0 and 3.x) is supported.

## Quick start

**Binary** - download an archive for your platform from the [releases](https://github.com/bropines/awg-wireproxy/releases):

```bash
tar xzf awg-wireproxy_<version>_linux_amd64.tar.gz
./wireproxy -c config.conf
```

**Docker** (multi-arch):

```bash
docker run -d --name awg-wireproxy --restart unless-stopped \
  -v "$PWD/config.conf:/etc/wireproxy/config:ro" -p 127.0.0.1:25344:25344 \
  ghcr.io/bropines/awg-wireproxy:latest
```

**Minimal config** (see [examples/awg.conf.example](examples/awg.conf.example) and the
[configuration guide](docs/configuration.md)):

```ini
[Interface]
Address = 10.200.200.2/32
PrivateKey = <your private key>
Jc = 5          # AmneziaWG options are optional
S1 = 20
S2 = 30

[Peer]
PublicKey = <server public key>
Endpoint = vpn.example.com:51820
AllowedIPs = 0.0.0.0/0

[Socks5]
BindAddress = 127.0.0.1:25344
```

Then point your browser or app at `socks5://127.0.0.1:25344`.

## Documentation

- 🚀 **[Installation & CLI usage](docs/usage.md)** - releases, build from source, command line flags.
- ⚙️ **[Configuration guide](docs/configuration.md)** - SOCKS5, HTTP, SNI, TCP/UDP tunnels, domain routing, multiple peers.
- 🛡️ **[AmneziaWG parameters](docs/amneziawg.md)** - `Jc`, `S1`-`S4`, `H1`-`H4`, `I1`-`I5` and the AWG 3.x options.
- 🖥️ **[Web panel](docs/admin-panel.md)** - status, config editor with validation, live logs, restart.
- 🐳 **[Docker guide](docs/docker.md)** - images, docker compose, running the panel in a container.
- 🏥 **[Health endpoints](docs/health-endpoint.md)** - `/metrics` & `/readyz`.
- 🦊 **[Use with VPN & browser extensions](UseWithVPN.md)** - Firefox container tabs and macOS auto-start.

**System services:** [systemd](systemd/systemd_docs.md) · [rc.d](rc.d/rc.d.md)

## Feature highlights

- WireGuard and AmneziaWG (1.0 - 3.x) with all obfuscation parameters
- SOCKS5 (with UDP) / HTTP(S) proxy and a transparent TLS (SNI) proxy
- Domain routing: send only matching destinations through the tunnel
- TCP static routing for client and server, UDP forwarding tunnel (`UDPProxyTunnel`), STDIO tunnel for `ssh -o ProxyCommand`
- Optional [web panel](docs/admin-panel.md) to inspect and reconfigure a running daemon
- Health endpoints for monitoring, multi-arch binaries and Docker images

## Why you might want this

- You simply want to use WireGuard/AWG as a way to proxy some traffic.
- You don't want root permission just to change WireGuard settings.
- You want to run it next to another VPN, or only for selected applications.

On **Android**, [WG Tunnel](https://github.com/wgtunnel/wgtunnel) (Proxy mode) covers the same use case with a
native app; this project targets Linux, macOS, Windows, BSD, servers and containers.

## Credits & authorship

This project is a heavily modified fork standing on the shoulders of giants:

- **[windtf/wireproxy](https://github.com/windtf/wireproxy)** (formerly octeep / pufferffish): the original creator and upstream; its features (SNI proxy, domain routing, UDP, health endpoints) are merged into this fork.
- **[artem-russkikh/wireproxy-awg](https://github.com/artem-russkikh/wireproxy-awg)**: added AmneziaWG (AWG) support.
- **[amnezia-vpn/amneziawg-go](https://github.com/amnezia-vpn/amneziawg-go)**: the AmneziaWG userspace implementation.
- **[bropines](https://github.com/bropines)**: current maintainer - UDP proxy tunnel, AWG 1.5-3.x support, web panel, build and release tooling.

## Stargazers over time

[![Stargazers over time](https://starchart.cc/bropines/awg-wireproxy.svg)](https://starchart.cc/bropines/awg-wireproxy)
