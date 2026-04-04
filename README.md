# wireproxy-awg

[![ISC licensed](https://img.shields.io/badge/license-ISC-blue)](./LICENSE)
[![Build status](https://github.com/bropines/awg-wireproxy/actions/workflows/build.yml/badge.svg)](https://github.com/bropines/wireproxy-awg/actions)

A wireguard and AmneziaWG client that exposes itself as a socks5/http proxy or tunnels.

## What is this

`wireproxy` is a completely userspace application that connects to a wireguard/AmneziaWG peer,
and exposes a socks5/http proxy or tunnels on the machine. This can be useful if you need
to connect to certain sites via a wireguard peer, but can't be bothered to setup a new network
interface for whatever reasons.

## Documentation

To keep things organized, the documentation is split into several modules. Please refer to the specific guides below:

- 🚀 **[Installation & CLI Usage](docs/usage.md)** - How to build, install, and run wireproxy.
- ⚙️ **[Configuration Guide](docs/configuration.md)** - Extensive configuration examples (SOCKS5, HTTP, TCP/UDP tunnels, AmneziaWG parameters).
- 🐳 **[Docker Guide](docs/docker.md)** - How to run wireproxy inside a Docker container.
- 🏥 **[Health Endpoints](docs/health-endpoint.md)** - Monitoring and health checks (`/metrics` & `/readyz`).
- 🦊 **[Use with VPN & Browser Extensions](UseWithVPN.md)** - Setup with Firefox Container Tabs and MacOS auto-start.

**System Services Integrations:**
- 🐧 [Running with systemd](systemd/README.md)
- 🧰 [Running with rc.d](rc.d/README.md)

## Feature Highlights

- TCP static routing for client and server
- UDP proxy and forwarding tunnel (`UDPProxyTunnel`)
- SOCKS5/HTTP proxy (currently only CONNECT is supported)
- Native AmneziaWG support

## Why you might want this

- You simply want to use wireguard/AWG as a way to proxy some traffic.
- You don't want root permission just to change wireguard settings.

Currently, I'm running wireproxy connected to a server in another country,
and configured my browser to use wireproxy for certain sites. It's pretty useful since
wireproxy is completely isolated from my network interfaces, and I don't need root to configure anything.

## Credits & Authorship

This project is a heavily modified fork standing on the shoulders of giants:
- **[windtf/wireproxy](https://github.com/windtf/wireproxy)** (formerly octeep): The original creator and upstream maintainer.
- **[artem-russkikh/wireproxy-awg](https://github.com/artem-russkikh/wireproxy-awg)**: Added AmneziaWG (AWG) support.
- **[bropines/wireproxy-awg](https://github.com/bropines/wireproxy-awg)**: Current maintainer. Added dynamic UDP Proxy Tunnel support, refined build systems, and overall structural improvements.

## Stargazers over time

[![Stargazers over time](https://starchart.cc/bropines/awg-wireproxy.svg)](https://starchart.cc/bropines/awg-wireproxy)