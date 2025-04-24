##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

## The build context is expected to be the root of the repository.
## Either build the image from the root of the repository:
##   docker build -f k8s/Dockerfile .
## Or build with docker-compose file of k8s

ARG BUILD_BASE_IMAGE=golang:1.22.4-alpine
ARG TARGET_BASE_IMAGE=mcr.microsoft.com/mirror/docker/library/alpine:3.16

FROM --platform=$BUILDPLATFORM ${BUILD_BASE_IMAGE} AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG BUILDFLAG=""

ENV CGO_ENABLED=0

ARG BUILD_BASE_IMAGE

ARG FAULT_INJECTION_ENABLED=false

# Install gcc, g++ and other necessary build tools
RUN if echo "${BUILD_BASE_IMAGE}" | grep "alpine"; then \
        apk add --no-cache gcc musl-dev curl; \
    elif echo "${BUILD_BASE_IMAGE}" | grep "mariner"; then \
        tdnf install -y gcc glibc-devel && tdnf clean all; \
    else \
        echo "Unsupported base image"; \
        exit 1; \
    fi

RUN go install github.com/magefile/mage@latest
WORKDIR /
# Copy the Go Modules 
COPY api/ api/
COPY coa/ coa/
COPY packages/ packages/
COPY k8s/ k8s/
WORKDIR /k8s
# Test
# Mariner images doesn't support TSan which is the required for go test -race.
RUN if echo "${BUILD_BASE_IMAGE}" | grep "mariner"; then \
        export SKIP_RACE_CHECK=true; \
        CGO_ENABLED=1 mage generate operatorTest; \
    else \
        CGO_ENABLED=1 mage generate operatorTest; \
    fi

# Install gofail
RUN if [ "$FAULT_INJECTION_ENABLED" = "true" ]; then \
    go install go.etcd.io/gofail@latest && \
    find /k8s -type d | while read -r dir; do gofail enable $dir; done && \
    cd /k8s && go get go.etcd.io/gofail/runtime; \
fi

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags "$BUILDFLAG" -o bin/manager
FROM ${TARGET_BASE_IMAGE} AS manager

WORKDIR /
COPY --from=builder /k8s/bin/manager .
USER 65532:65532
ENV GOFAIL_HTTP="127.0.0.1:22381"

ENTRYPOINT ["/manager"]