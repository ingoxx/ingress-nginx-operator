# Build the manager binary
#FROM golang:1.21 AS builder
FROM gotec007/go:v1.22 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY external/ external/
COPY rootfs/ rootfs/
COPY pkg/ pkg/
COPY utils/ utils/
COPY services/ services/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager main.go && \
    go build -a -o httpserver utils/http/httpserver.go

FROM nginx:latest AS chroot

WORKDIR /

COPY --from=builder /workspace/rootfs /rootfs
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/httpserver .
RUN chown -R www-data:www-data /rootfs manager && \
    chmod +x rootfs/etc/nginx/shell/start.sh && \
    chmod +x /manager && \
    chmod +x /httpserver


# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# FROM gcr.io/distroless/static:nonroot
FROM gotec007/nginx:noroot5.3

WORKDIR /

COPY --from=chroot /manager .
COPY --from=chroot /httpserver .
COPY --from=chroot /rootfs /rootfs

USER 33:33

ENTRYPOINT ["/manager"]

