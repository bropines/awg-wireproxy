# Running in Docker

Wireproxy can easily be deployed using Docker. This avoids installing dependencies on your host machine and integrates cleanly into existing containerized environments.

## Pulling the Image

If you use the pre-built images from the GitHub Container Registry:

```bash
docker pull ghcr.io/bropines/awg-wireproxy:latest
```

## Running the Container

Create your configuration file (e.g., `wireproxy.conf`) on your host machine.

Run the container, mapping your configuration file into the container and publishing any proxy ports you specified in your config:

```bash
docker run -d \
  --name awg-wireproxy \
  --restart unless-stopped \
  -v /path/to/your/wireproxy.conf:/etc/wireproxy/config:ro \
  -p 1080:1080 \
  ghcr.io/bropines/awg-wireproxy:latest
```
*(Replace `-p 1080:1080` with the actual `BindAddress` ports defined in your `[Socks5]`, `[http]`, or `[TCPClientTunnel]` configurations).*

## Building the Image Locally

If you prefer to build the image yourself:

```bash
git clone https://github.com/bropines/awg-wireproxy.git
cd awg-wireproxy
docker build -t awg-wireproxy-local .
```

You can then run your local image just like the remote one:

```bash
docker run -d \
  -v $(pwd)/myconfig.conf:/etc/wireproxy/config:ro \
  -p 25344:25344 \
  awg-wireproxy-local
```