# Based on https://actuated.dev/blog/multi-arch-docker-github-actions

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.20-alpine as builder
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /go/src/github.com/paulfantom/cats

COPY .  .

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -ldflags "-s -w" \
  -a -o /usr/bin/cats .

#FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static:nonroot
FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:3.18.0

LABEL org.opencontainers.image.source=https://github.com/paulfantom/cats

WORKDIR /
COPY --from=builder /usr/bin/cats /
#USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/cats"]