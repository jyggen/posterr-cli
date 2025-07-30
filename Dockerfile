FROM --platform=$BUILDPLATFORM cgr.dev/chainguard/go:latest@sha256:2035c1a00329b02640d83bdb68306ff3f6d40316b1a0cd74e11554f0830a46fa AS builder
ARG TARGETOS
ARG TARGETARCH
COPY . /app
RUN cd /app && GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -trimpath -o posterr cmd/posterr/main.go

FROM cgr.dev/chainguard/static:latest
COPY --from=builder /app/posterr /usr/bin/
ENTRYPOINT ["/usr/bin/posterr"]