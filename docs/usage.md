# Installation & Usage

## Prebuilt binaries

Every [release](https://github.com/bropines/awg-wireproxy/releases) contains archives for

| OS | Architectures |
|---|---|
| Linux | amd64, arm64, armv6, armv7, 386, mips, mipsle (softfloat), riscv64, s390x, ppc64le |
| Windows | amd64, arm64, 386 |
| macOS | amd64 (Intel), arm64 (Apple Silicon) |
| FreeBSD | amd64, arm64 |

and a `checksums.txt` to verify them.

```bash
tar xzf awg-wireproxy_<version>_linux_amd64.tar.gz
./wireproxy --version
./wireproxy -c config.conf
```

The archives also contain the license, `examples/` and the systemd unit.

## Docker

```bash
docker pull ghcr.io/bropines/awg-wireproxy:latest
```

See the [Docker guide](docker.md) (includes docker compose and the web panel).

## Build from source

You need Go (the version in `go.mod`).

```bash
git clone https://github.com/bropines/awg-wireproxy.git
cd awg-wireproxy
make          # binary: build/wireproxy, version stamped from git
make test     # go vet + go test -race (incl. loopback WireGuard/AmneziaWG tunnel tests)
```

Check the `Makefile` for more targets (`make build-all` cross-compiles for Linux, Windows and macOS,
`make android-bin`, `make aar`).

Or install straight from the repository:

```bash
go install github.com/bropines/awg-wireproxy/cmd/wireproxy@master
```

## CLI usage

```bash
./wireproxy -c /path/to/config
```

```text
usage: wireproxy [-h|--help] [-c|--config "<value>"] [-s|--silent]
                 [-d|--daemon] [-i|--info "<value>"] [-v|--version]
                 [-n|--configtest] [-a|--admin "<value>"] [-t|--admin-token
                 "<value>"]

Arguments:

  -h  --help         Print help information
  -c  --config       Path of configuration file
                     Default paths: /etc/wireproxy/wireproxy.conf, $HOME/.config/wireproxy.conf
  -s  --silent       Silent mode
  -d  --daemon       Make wireproxy run in background
  -i  --info         Specify the address and port for exposing health status
  -v  --version      Print version
  -n  --configtest   Configtest mode. Only check the configuration file for validity.
  -a  --admin        Enable the web panel on this address (e.g. 127.0.0.1:9090).
                     Env: WIREPROXY_ADMIN. Disables the sandbox
  -t  --admin-token  Access token for the web panel (random if empty).
                     Env: WIREPROXY_ADMIN_TOKEN
```

- `-n` validates a config (value ranges, AmneziaWG rules, regexes, hostnames) without starting anything.
- `-a` turns on the [web panel](admin-panel.md): status, config editor, live logs and restart.
- `-i` exposes the [health endpoints](health-endpoint.md).
