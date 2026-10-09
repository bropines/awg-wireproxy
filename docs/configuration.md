# Configuration Guide

The `wireproxy` configuration file combines standard WireGuard/AmneziaWG syntax with custom routing blocks.

## Sample Config File

```ini
# The [Interface] and [Peer] configurations follow the same semantics and meaning
# of a wg-quick configuration.
[Interface]
Address = 10.200.200.2/32 # The subnet should be /32 and /128 for IPv4 and v6 respectively
# MTU = 1420 (optional)
PrivateKey = uCTIK+56CPyCvwJxmU5dBfuyJvPuSXAq1FzHdnIxe1Q=
# PrivateKey = $MY_WIREGUARD_PRIVATE_KEY # Alternatively, reference environment variables
DNS = 10.200.200.1

# AmneziaWG fields (Jc, S1-S4, H1-H4, I1-I5 and the AWG 3.x options) go here too,
# see amneziawg.md for the full list.
# Jc = 4
# Jmin = 50
# Jmax = 1000
# S1 = 40
# S2 = 40
# H1 = 1
# H2 = 2
# H3 = 3
# H4 = 4

[Peer]
PublicKey = QP+A67Z2UBrMgvNIdHv8gPel5URWNLS4B3ZQ2hQIZlg=
# PresharedKey = UItQuvLsyh50ucXHfjF0bbR4IIpVBd74lwKc8uIPXXs= (optional)
Endpoint = my.ddns.example.com:51820
# PersistentKeepalive = 25 (optional)

# TCPClientTunnel is a tunnel listening on your machine,
# and it forwards any TCP traffic received to the specified target via wireguard.
# Flow: <an app on your LAN> --> localhost:25565 --(wireguard)--> play.cubecraft.net:25565
[TCPClientTunnel]
BindAddress = 127.0.0.1:25565
Target = play.cubecraft.net:25565

# TCPServerTunnel is a tunnel listening on wireguard,
# and it forwards any TCP traffic received to the specified target via local network.
# Flow: <an app on your wg network> --(wireguard)--> 172.16.31.2:3422 --> localhost:25545
[TCPServerTunnel]
ListenPort = 3422
Target = localhost:25545

# UDPProxyTunnel listens locally and forwards dynamic UDP traffic via wireguard.
# Flow: <an app on your LAN> --> localhost:53 --(wireguard)--> 1.1.1.1:53
[UDPProxyTunnel]
BindAddress = 127.0.0.1:53
Target = 1.1.1.1:53
InactivityTimeout = 30 # If its set to 0, it will never timeout

# STDIOTunnel connects standard input and output to the specified TCP target.
# Useful for ProxyCommand in ssh (e.g. ssh -o ProxyCommand='wireproxy -c config.conf' myserver)
[STDIOTunnel]
Target = ssh.myserver.net:22

# Socks5 creates a socks5 proxy on your LAN
[Socks5]
BindAddress = 127.0.0.1:25344
#Username = ...
#Password = ...
# Route only matching domains through wireguard, see "Domain routing" below.
#TunnelDomains = ^(.*\.)?example\.com$

# http creates a http proxy on your LAN
[http]
BindAddress = 127.0.0.1:25345
#Username = ...
#Password = ...
#CertFile = ...
#KeyFile = ...

# SNI creates a transparent TLS proxy: the destination is taken from the TLS
# Server Name Indication and the connection is routed via wireguard.
[SNI]
BindAddress = 0.0.0.0:443
```

## AmneziaWG

Every AmneziaWG option (1.0 to 3.x) is written in `[Interface]`; see the [AmneziaWG parameters](amneziawg.md).

## Domain routing (`TunnelDomains`)

`[Socks5]`, `[http]` and `[SNI]` can send only some destinations through wireguard and dial everything else
directly over your normal network:

```ini
[Socks5]
BindAddress = 127.0.0.1:25344
# Each TunnelDomains line is one full Go (RE2) regular expression. Repeat the key for several patterns;
# do NOT comma-separate (quantifiers like {2,4} keep working). Matching is case-insensitive and a
# trailing dot is ignored. Without TunnelDomains everything is routed through wireguard (default).
TunnelDomains = ^(.*\.)?example\.com$
TunnelDomains = ^ipinfo\.io$
# Log every connection's destination and whether it went to the TUNNEL or DIRECT. Useful to discover
# which domains your apps reach before writing TunnelDomains.
LogDomains = true
```

For `[SNI]` the patterns are matched against the TLS SNI host name.

## Importing Existing WG Config

If you already have a `.conf` file, you can import it:

```ini
WGConfig = /path/to/existing/wg.conf

[TCPClientTunnel]
BindAddress = 127.0.0.1:25565
Target = play.cubecraft.net:25565

[Socks5]
BindAddress = 127.0.0.1:25344
```

## Multiple Peers & DNS Resolution

Having multiple peers is supported. `AllowedIPs` determines which peer traffic is routed to.

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

[Resolve]
# Set DNS Resovle Strategy
# `ipv4`: Prioritize A records.
# `ipv6`: Prioritize AAAA records.
# `auto` (Default): Equivalent to ipv4 if WG interface has IPv4 only, else ipv6.
ResolveStrategy = auto 
```

Wireproxy can also allow peers to connect to it (act as a server) by specifying `ListenPort` in `[Interface]` and omitting `Endpoint` in `[Peer]`.

## Managing the config from the browser

Start wireproxy with `--admin 127.0.0.1:9090` to get a [web panel](admin-panel.md) that edits and validates this file
and restarts the daemon for you.