# syntax=docker/dockerfile:1

# Build natively on the CI machine and cross-compile for the target platform,
# so multi-arch images do not need slow QEMU emulation.
FROM --platform=$BUILDPLATFORM docker.io/golang:1.27 AS build
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG VERSION=dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} \
    go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/wireproxy ./cmd/wireproxy

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/wireproxy /usr/bin/wireproxy

# The config lives here. To use the web panel, mount a directory the
# container user (uid 65532) can write to.
VOLUME [ "/etc/wireproxy" ]
ENTRYPOINT [ "/usr/bin/wireproxy" ]
CMD [ "--config", "/etc/wireproxy/config" ]

LABEL org.opencontainers.image.title="awg-wireproxy"
LABEL org.opencontainers.image.description="WireGuard / AmneziaWG client that exposes itself as a socks5/http proxy, with an optional web panel"
LABEL org.opencontainers.image.source="https://github.com/bropines/awg-wireproxy"
LABEL org.opencontainers.image.licenses="ISC"
