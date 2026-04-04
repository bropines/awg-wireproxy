# Start by building the application.
FROM docker.io/golang:1.26 as build

WORKDIR /usr/src/wireproxy
COPY . .

RUN CGO_ENABLED=0 make

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian11:nonroot
COPY --from=build /usr/src/wireproxy/build/wireproxy /usr/bin/wireproxy

VOLUME [ "/etc/wireproxy"]
ENTRYPOINT [ "/usr/bin/wireproxy" ]
CMD [ "--config", "/etc/wireproxy/config" ]

LABEL org.opencontainers.image.title="wireproxy-awg"
LABEL org.opencontainers.image.description="AmneziaWG/Wireguard client that exposes itself as a socks5/http proxy and UDP/TCP tunnels"
LABEL org.opencontainers.image.licenses="ISC"
LABEL org.opencontainers.image.source="https://github.com/bropines/wireproxy-awg"