# Running in Docker

Wireproxy runs well in a container: no dependencies on the host, and the image is a ~5 MB distroless
image running as a non-root user. Multi-arch images (amd64, arm64, arm/v7, ppc64le, s390x) are published
to the GitHub Container Registry.

| Tag | Meaning |
|---|---|
| `latest` | newest stable release |
| `X.Y.Z`, `X.Y` | a specific release / the newest patch of a minor version |
| `edge` | latest build of the default branch |

```bash
docker pull ghcr.io/bropines/awg-wireproxy:latest
```

## Plain proxy

Create a config (start from [examples/awg.conf.example](../examples/awg.conf.example)) and mount it as
`/etc/wireproxy/config`. Inside a container bind the proxies to `0.0.0.0` and publish the ports:

```bash
docker run -d \
  --name awg-wireproxy \
  --restart unless-stopped \
  -v /path/to/wireproxy.conf:/etc/wireproxy/config:ro \
  -p 127.0.0.1:25344:25344 \
  ghcr.io/bropines/awg-wireproxy:latest
```

Replace the `-p` mapping with the `BindAddress` ports from your `[Socks5]`, `[http]` or `[TCPClientTunnel]`
sections. In this mode the process keeps its landlock sandbox.

## With the web panel (docker compose)

The repository ships a [docker-compose.yml](../docker-compose.yml):

```bash
mkdir config && cp examples/awg.conf.example config/config   # or create the config later in the panel
sudo chown -R 65532:65532 config                             # the image runs as uid 65532
echo "ADMIN_TOKEN=$(openssl rand -hex 24)" > .env
docker compose up -d
```

Open <http://127.0.0.1:9090> and sign in with the token. The config directory is mounted **read-write**
because the panel saves the config (and a `.bak` of the previous one) there; applying a change restarts the
process inside the container, the container itself keeps running. See the [web panel guide](admin-panel.md)
for what it can do and its security notes. Do not publish port 9090 to the internet without a TLS reverse proxy.

## Building the image locally

```bash
git clone https://github.com/bropines/awg-wireproxy.git
cd awg-wireproxy
docker build -t awg-wireproxy-local .
docker run -d -v $(pwd)/myconfig.conf:/etc/wireproxy/config:ro -p 25344:25344 awg-wireproxy-local
```

The Dockerfile cross-compiles on the build machine, so `docker buildx build --platform linux/arm64 .`
does not need QEMU.
