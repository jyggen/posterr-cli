FROM --platform=$BUILDPLATFORM cgr.dev/chainguard/go:latest@sha256:2035c1a00329b02640d83bdb68306ff3f6d40316b1a0cd74e11554f0830a46fa AS builder
ARG TARGETOS
ARG TARGETARCH
COPY . /app
RUN cd /app && GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -trimpath -o posterr cmd/posterr/main.go

FROM cgr.dev/chainguard/static:latest@sha256:81b61e16687f76ebc3c1fa71ec3fa3e0901e2908e0cd442f378557c294920aac
COPY --from=builder /app/posterr /usr/bin/
ENTRYPOINT ["/usr/bin/posterr"]