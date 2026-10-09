# Health Endpoints

Wireproxy supports exposing a health endpoint for monitoring purposes.
The argument `--info/-i` specifies an address and port (e.g. `localhost:9080`), which exposes a HTTP server providing health status metrics.

Currently two endpoints are implemented:

## `/metrics`
Exposes information of the wireguard daemon, providing the same information you would get with `wg show`.
[This](https://www.wireguard.com/xplatform/#example-dialog) shows an example of what the response looks like. *(Note: private and preshared keys are redacted for security).*

## `/readyz`
Responds with JSON showing the last time a pong was received from an IP specified with `CheckAlive`.

When `CheckAlive` is set in the config, a ping is sent out to addresses in `CheckAlive` every `CheckAliveInterval` seconds (defaults to 5) via wireguard.
If a pong has not been received within `CheckAliveInterval` seconds (+2 seconds leeway for latency), it responds with HTTP `503 Service Unavailable`, otherwise `200 OK`.

### Example Config

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
```

### Example Responses

If `3.3.3.3` is unreachable, `/readyz` would respond with `503`:

```text
< HTTP/1.1 503 Service Unavailable
< Content-Type: text/plain; charset=utf-8
<
{"1.1.1.1":1712796899,"3.3.3.3":0}
```

If all IPs are reachable (or if only `1.1.1.1` was configured), it responds with `200`:

```text
< HTTP/1.1 200 OK
< Content-Type: text/plain; charset=utf-8
<
{"1.1.1.1":1712796979}
```

*Note: If nothing is set for `CheckAlive`, an empty JSON object `{}` with `200 OK` will be returned.*