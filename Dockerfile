FROM cgr.dev/chainguard/static:latest@sha256:7d8e6efa03a7b58b5a5b2a1d8555e44b990775b29d6324e12d1c77314d595aaa
ENTRYPOINT ["/usr/bin/posterr"]
COPY posterr /usr/bin/posterr
