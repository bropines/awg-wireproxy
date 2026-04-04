# Start by building the application.
FROM docker.io/golang:1.23-alpine AS build

WORKDIR /usr/src/wireproxy
COPY . .

RUN apk add --no-cache make git

RUN CGO_ENABLED=0 make

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian11:nonroot
COPY --from=build /usr/src/wireproxy/build/wireproxy /usr/bin/wireproxy

VOLUME [ "/etc/wireproxy" ]
ENTRYPOINT [ "/usr/bin/wireproxy" ]
CMD [ "-c", "/etc/wireproxy/config" ]

LABEL org.opencontainers.image.title="awg-wireproxy"
LABEL org.opencontainers.image.description="AmneziaWG/Wireguard client that exposes itself as a socks5/http proxy and UDP/TCP tunnels"
LABEL org.opencontainers.image.licenses="ISC"
LABEL org.opencontainers.image.source="https://github.com/bropines/awg-wireproxy"