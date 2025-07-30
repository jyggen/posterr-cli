FROM --platform=$BUILDPLATFORM cgr.dev/chainguard/go:latest@sha256:98f013454e586ce641e193214930620f092081a5ca19275e96b9599e97c3ae7a AS builder
ARG TARGETOS
ARG TARGETARCH
COPY . /app
RUN cd /app && GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -trimpath -o posterr cmd/posterr/main.go

FROM cgr.dev/chainguard/static:latest@sha256:81b61e16687f76ebc3c1fa71ec3fa3e0901e2908e0cd442f378557c294920aac
COPY --from=builder /app/posterr /usr/bin/
ENTRYPOINT ["/usr/bin/posterr"]