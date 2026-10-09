# AmneziaWG parameters

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

See the [configuration guide](configuration.md) for the rest of the config file.
