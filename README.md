# awg-wireproxy

[![MIT/ISC licensed](https://img.shields.io/badge/license-ISC-blue)](./LICENSE)
[![Build status](https://github.com/bropines/awg-wireproxy/actions/workflows/build.yml/badge.svg)](https://github.com/bropines/awg-wireproxy/actions)

A WireGuard **and AmneziaWG** client that exposes itself as a socks5/http proxy or tunnels.

This is a fork of [windtf/wireproxy](https://github.com/windtf/wireproxy) whose core is
[amneziawg-go](https://github.com/amnezia-vpn/amneziawg-go) **v3.1** instead of wireguard-go.
It stays in sync with upstream (SNI proxy, domain routing, UDP, health endpoints) and adds full
AmneziaWG support: AWG 1.0, 1.5 (`I1`-`I5`), 2.0 (`S3`/`S4`, `H1`-`H4` ranges) and 3.x
(header protection, content padding, custom timings). A plain WireGuard config works unchanged.

# What is this

`wireproxy` is a completely userspace application that connects to a wireguard (or AmneziaWG) peer,
and exposes a socks5/http proxy or tunnels on the machine. This can be useful if you need
to connect to certain sites via a wireguard peer, but can't be bothered to setup a new network
interface for whatever reasons.

# Why you might want this

- You simply want to use wireguard as a way to proxy some traffic.
- You don't want root permission just to change wireguard settings.

Currently, I'm running wireproxy connected to a wireguard server in another country,
and configured my browser to use wireproxy for certain sites. It's pretty useful since
wireproxy is completely isolated from my network interfaces, and I don't need root to configure
anything.

Users who want something similar but for Amnezia VPN can use [this fork](https://github.com/artem-russkikh/wireproxy-awg)
of wireproxy by [@artem-russkikh](https://github.com/artem-russkikh).

# Feature

- WireGuard and AmneziaWG (1.0 - 3.x) with all obfuscation parameters
- TCP static routing for client and server
- SOCKS5/HTTP proxy (currently only CONNECT is supported)
- Transparent TLS ([SNI](https://en.wikipedia.org/wiki/Server_Name_Indication)) proxy

# TODO

- UDP Support in SOCKS5
- UDP static routing

# Usage

```bash
./wireproxy [-c path to config]
```

```bash
usage: wireproxy [-h|--help] [-c|--config "<value>"] [-s|--silent]
                 [-d|--daemon] [-i|--info "<value>"] [-v|--version]
                 [-n|--configtest]

                 Userspace wireguard client for proxying

Arguments:

  -h  --help        Print help information
  -c  --config      Path of configuration file
                    Default paths: /etc/wireproxy/wireproxy.conf, $HOME/.config/wireproxy.conf
  -s  --silent      Silent mode
  -d  --daemon      Make wireproxy run in background
  -i  --info        Specify the address and port for exposing health status
  -v  --version     Print version
  -n  --configtest  Configtest mode. Only check the configuration file for
                    validity.
```

# Build instruction

```bash
git clone https://github.com/bropines/awg-wireproxy
cd awg-wireproxy
make        # binary: build/wireproxy
```

# Install

```bash
go install github.com/bropines/awg-wireproxy/cmd/wireproxy@latest
```

# AmneziaWG

Put the AmneziaWG parameters into the `[Interface]` section (or into the file referenced by `WGConfig`),
exactly as in an `awg-quick` config. Every parameter is optional; omitted ones keep the plain WireGuard
behavior. Keys are case-insensitive.

```ini
[Interface]
Address = 10.200.200.2/32
PrivateKey = uCTIK+56CPyCvwJxmU5dBfuyJvPuSXAq1FzHdnIxe1Q=

# AWG 1.0 - junk packets before the handshake and padding of handshake messages
Jc = 5
Jmin = 10
Jmax = 50
S1 = 20
S2 = 30
# AWG 2.0 - padding of cookie / transport messages, magic headers (single value or range)
S3 = 40
S4 = 50
H1 = 100000-200000
H2 = 300000-400000
H3 = 500000-600000
H4 = 700000-800000
# AWG 1.5 - signature ("protocol mimicry") packets sent before the handshake
I1 = <b 0xc7000000010800000000000000000000><r 32><t>
# I2 .. I5 work the same way

# AWG 3.x
HeaderProtectionKey = <32-byte key, base64 (awg genkey) or 64 hex chars>
ContentPaddingAddition = 4-16
RekeyAfterTime = 100-120
RekeyTimeout = 4-6
RejectAfterTime = 170-190
KeepaliveTimeout = 8-12
MaxHandshakeAttempts = 5
RandomTrailers = true
DisableCookies = false

[Peer]
PublicKey = QP+A67Z2UBrMgvNIdHv8gPel5URWNLS4B3ZQ2hQIZlg=
Endpoint = my.ddns.example.com:51820
PersistentKeepalive = 20-30   # AWG 3: a range is accepted, plain seconds still work
```

| Key | Version | Meaning | Must match the server |
|---|---|---|---|
| `Jc`, `Jmin`, `Jmax` | 1.0 | Number of junk packets (1-128) and their size range (`Jmin` <= `Jmax` <= 1280) | no |
| `S1`, `S2` | 1.0 | Random padding added to the handshake init / response | yes |
| `S3`, `S4` | 2.0 | Padding of cookie-reply / transport messages | yes |
| `H1` - `H4` | 1.0 / 2.0 | Magic headers of init / response / cookie / transport. A single value (1.0) or a range `a-b` (2.0); the four ranges must not overlap | yes |
| `I1` - `I5` | 1.5 | Signature packets that mimic another protocol. Tags: `<b 0x..>` static bytes, `<r N>` random bytes, `<rd N>` random digits, `<rc N>` random chars, `<t>` timestamp, `<c>` packet counter | no |
| `HeaderProtectionKey` | 3.x | Key used to encrypt low-entropy header fields. Requires `S1`-`S4` >= 12 | yes |
| `ContentPaddingAddition` | 3.x | Extra random padding range for data packets | recommended |
| `RekeyAfterTime`, `RekeyTimeout`, `RejectAfterTime`, `KeepaliveTimeout` | 3.x | Override WireGuard timings, in seconds (value or range) | no |
| `MaxHandshakeAttempts` | 3.x | Maximum handshake retries (value or range) | no |
| `RandomTrailers` | 3.x | Append random trailers to packets | no |
| `DisableCookies` | 3.x | Disable cookie replies | no |
| `PersistentKeepalive` (`[Peer]`) | 3.x | Seconds or a range `a-b` | no |

`wireproxy -n -c config.conf` validates the file (value ranges, `S1`-`S4` sizes, overlapping `H` ranges, the
`HeaderProtectionKey` requirements) without starting the tunnel.

If you only need a SOCKS5/HTTP proxy **on Android**, [WG Tunnel](https://github.com/wgtunnel/wgtunnel)
(Proxy mode) is a better choice; this project targets Linux, macOS, Windows, servers and Docker.

# Use with VPN

Instructions for using wireproxy with Firefox container tabs and auto-start on MacOS can be found [here](/UseWithVPN.md).

# Sample config file

```ini
# The [Interface] and [Peer] configurations follow the same semantics and meaning
# of a wg-quick configuration. To understand what these fields mean, please refer to:
# https://wiki.archlinux.org/title/WireGuard#Persistent_configuration
# https://www.wireguard.com/#simple-network-interface
[Interface]
Address = 10.200.200.2/32 # The subnet should be /32 and /128 for IPv4 and v6 respectively
# MTU = 1420 (optional)
PrivateKey = uCTIK+56CPyCvwJxmU5dBfuyJvPuSXAq1FzHdnIxe1Q=
# PrivateKey = $MY_WIREGUARD_PRIVATE_KEY # Alternatively, reference environment variables
DNS = 10.200.200.1

[Peer]
PublicKey = QP+A67Z2UBrMgvNIdHv8gPel5URWNLS4B3ZQ2hQIZlg=
# PresharedKey = UItQuvLsyh50ucXHfjF0bbR4IIpVBd74lwKc8uIPXXs= (optional)
Endpoint = my.ddns.example.com:51820
# PersistentKeepalive = 25 (optional)

# TCPClientTunnel is a tunnel listening on your machine,
# and it forwards any TCP traffic received to the specified target via wireguard.
# Flow:
# <an app on your LAN> --> localhost:25565 --(wireguard)--> play.cubecraft.net:25565
[TCPClientTunnel]
BindAddress = 127.0.0.1:25565
Target = play.cubecraft.net:25565

# TCPServerTunnel is a tunnel listening on wireguard,
# and it forwards any TCP traffic received to the specified target via local network.
# Flow:
# <an app on your wireguard network> --(wireguard)--> 172.16.31.2:3422 --> localhost:25545
[TCPServerTunnel]
ListenPort = 3422
Target = localhost:25545

# STDIOTunnel is a tunnel connecting the standard input and output of the wireproxy
# process to the specified TCP target via wireguard.
# This is especially useful to use wireproxy as a ProxyCommand parameter in openssh
# For example:
#    ssh -o ProxyCommand='wireproxy -c myconfig.conf' ssh.myserver.net
# Flow:
# Piped command -->(wireguard)--> ssh.myserver.net:22
[STDIOTunnel]
Target = ssh.myserver.net:22

# Socks5 creates a socks5 proxy on your LAN, and all traffic would be routed via wireguard.
[Socks5]
BindAddress = 127.0.0.1:25344

# Socks5 authentication parameters, specifying username and password enables
# proxy authentication.
#Username = ...
# Avoid using spaces in the password field
#Password = ...

# Domain whitelist routing (optional). When TunnelDomains is set, only connections
# whose destination host matches one of the patterns are routed through wireguard;
# every other connection is dialed directly over your normal network. When
# TunnelDomains is unset, all traffic is routed through wireguard (default).
# Each TunnelDomains line is a single, full Go regular expression (RE2). Repeat
# the key for multiple patterns; do NOT comma-separate (so quantifiers like {2,4}
# keep working). Matching is case-insensitive and a trailing dot is ignored.
#TunnelDomains = ^(.*\.)?example\.com$
#TunnelDomains = ^ipinfo\.io$
# Set LogDomains = true to log every connection's destination host and whether it
# was routed to the TUNNEL or DIRECT. Useful for discovering which domains your
# apps reach before writing TunnelDomains. Off by default.
#LogDomains = true

# http creates a http proxy on your LAN, and all traffic would be routed via wireguard.
[http]
BindAddress = 127.0.0.1:25345

# HTTP authentication parameters, specifying username and password enables
# proxy authentication.
#Username = ...
# Avoid using spaces in the password field
#Password = ...

# Specifying certificate and key enables HTTPS
#CertFile = ...
#KeyFile = ...

# TunnelDomains / LogDomains work here too (same semantics as [Socks5] above).
#TunnelDomains = ^(.*\.)?example\.com$
#LogDomains = true

# SNI creates a transparent TLS proxy on your LAN, and all traffic would be routed via wireguard,
# using Server Name Indication as routing destination.
[SNI]
BindAddress = 0.0.0.0:443

# TunnelDomains / LogDomains work here too, matched against the TLS SNI hostname.
#TunnelDomains = ^(.*\.)?example\.com$
#LogDomains = true
```

Alternatively, if you already have a wireguard config, you can import it in the
wireproxy config file like this:

```ini
WGConfig = <path to the wireguard config>

# Same semantics as above
[TCPClientTunnel]
...

[TCPServerTunnel]
...

[Socks5]
...
```

Having multiple peers is also supported. `AllowedIPs` would need to be specified
such that wireproxy would know which peer to forward to.

```ini
[Interface]
Address = 10.254.254.40/32
PrivateKey = XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX=

[Peer]
Endpoint = 192.168.0.204:51820
PublicKey = YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY=
AllowedIPs = 10.254.254.100/32
PersistentKeepalive = 25

[Peer]
PublicKey = ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ=
AllowedIPs = 10.254.254.1/32, fdee:1337:c000:d00d::1/128
Endpoint = 172.16.0.185:44044
PersistentKeepalive = 25


[TCPServerTunnel]
ListenPort = 5000
Target = service-one.servicenet:5000

[TCPServerTunnel]
ListenPort = 5001
Target = service-two.servicenet:5001

[TCPServerTunnel]
ListenPort = 5080
Target = service-three.servicenet:80

[UDPProxyTunnel]
BindAddress = 127.0.0.1:53
Target = 1.1.1.1:53
InactivityTimeout = 30 # If its set to 0, it will never timeout

[Resolve]
# Set DNS Resovle Strategy
# `ipv4`: Prioritize A records.
# `ipv6`: Prioritize AAAA records       .
# `auto` (Default): If the WireGuard interface has IPv4 address only, it's equivalent to `ipv4`, otherwise it's equivalent to `ipv6`.
ResolveStrategy = auto 
```

Wireproxy can also allow peers to connect to it:

```ini
[Interface]
ListenPort = 5400
...

[Peer]
PublicKey = YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY=
AllowedIPs = 10.254.254.100/32
# Note there is no Endpoint defined here.
```

# Health endpoint

Wireproxy supports exposing a health endpoint for monitoring purposes.
The argument `--info/-i` specifies an address and port (e.g. `localhost:9080`), which exposes a HTTP server that provides health status metric of the server.

Currently two endpoints are implemented:

`/metrics`: Exposes information of the wireguard daemon, this provides the same information you would get with `wg show`. [This](https://www.wireguard.com/xplatform/#example-dialog) shows an example of what the response would look like.

`/readyz`: This responds with a json which shows the last time a pong is received from an IP specified with `CheckAlive`. When `CheckAlive` is set, a ping is sent out to addresses in `CheckAlive` per `CheckAliveInterval` seconds (defaults to 5) via wireguard. If a pong has not been received from one of the addresses within the last `CheckAliveInterval` seconds (+2 seconds for some leeway to account for latency), then it would respond with a 503, otherwise a 200.

For example:

```ini
[Interface]
PrivateKey = censored
Address = 10.2.0.2/32
DNS = 10.2.0.1
CheckAlive = 1.1.1.1, 3.3.3.3
CheckAliveInterval = 3

[Peer]
PublicKey = censored
AllowedIPs = 0.0.0.0/0
Endpoint = 149.34.244.174:51820

[Socks5]
BindAddress = 127.0.0.1:25344
```

`/readyz` would respond with

```text
< HTTP/1.1 503 Service Unavailable
< Date: Thu, 11 Apr 2024 00:54:59 GMT
< Content-Length: 35
< Content-Type: text/plain; charset=utf-8
<
{"1.1.1.1":1712796899,"3.3.3.3":0}
```

And for:

```ini
[Interface]
PrivateKey = censored
Address = 10.2.0.2/32
DNS = 10.2.0.1
CheckAlive = 1.1.1.1
```

`/readyz` would respond with

```text
< HTTP/1.1 200 OK
< Date: Thu, 11 Apr 2024 00:56:21 GMT
< Content-Length: 23
< Content-Type: text/plain; charset=utf-8
<
{"1.1.1.1":1712796979}
```

If nothing is set for `CheckAlive`, an empty JSON object with 200 will be the response.

The peer which the ICMP ping packet is routed to depends on the `AllowedIPs` set for each peers.


# Credits

- [windtf/wireproxy](https://github.com/windtf/wireproxy) (originally [pufferffish/wireproxy](https://github.com/pufferffish/wireproxy)) - the base of this project
- [amnezia-vpn/amneziawg-go](https://github.com/amnezia-vpn/amneziawg-go) - the AmneziaWG userspace implementation
- [artem-russkikh/wireproxy-awg](https://github.com/artem-russkikh/wireproxy-awg) - the original AmneziaWG port of wireproxy
