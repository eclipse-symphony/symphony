##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

ARG BUILD_BASE_IMAGE=ubuntu:22.04
ARG TARGET_BASE_IMAGE=ubuntu:latest

FROM --platform=$BUILDPLATFORM ${BUILD_BASE_IMAGE} AS build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG FAULT_INJECTION_ENABLED=false
ARG BUILDFLAG=""

ENV CGO_ENABLED=1
ENV GO_VERSION=1.22.4
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir /workspace /extensions
COPY ./packages /workspace/packages
COPY ./coa /workspace/coa
COPY ./api /workspace/api
COPY ./remote-agent /workspace/remote-agent

WORKDIR /workspace/remote-agent
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=linux GOARCH=amd64 go build -o /remote-agent/remote-agent
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=windows GOARCH=amd64 go build -o /remote-agent/remote-agent.exe

WORKDIR /workspace/api
# File permissions are not preserved when copying files in ADO. 
RUN chmod +x pkg/apis/v1alpha1/providers/target/script/mock-*.sh

# Install necessary build tools, cross-compilers, and Go
RUN if [ "$BUILDFLAG" != "azure" ]; then \
    apt-get update && apt-get install -y \
    gcc g++ make curl wget git cargo \
    gcc-arm-linux-gnueabihf gcc-aarch64-linux-gnu \
    libssl-dev pkg-config \
    && wget https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz; \
    fi

# Install Rust and rustup, overriding any existing Rust installation
RUN if [ "$BUILDFLAG" != "azure" ]; then \
    curl https://sh.rustup.rs -sSf | sh -s -- -y --no-modify-path \
    && . "$HOME/.cargo/env" \
    && rustup target add armv7-unknown-linux-gnueabihf aarch64-unknown-linux-gnu; \
    fi
# Set up Rust environment
ENV CARGO_TARGET_ARM_UNKNOWN_LINUX_GNUEABIHF_LINKER=arm-linux-gnueabihf-gcc
ENV CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc
ENV PATH=/root/.cargo/bin:$PATH

# Build the Rust component
WORKDIR /workspace/api/pkg/apis/v1alpha1/providers/target/rust

# Clean build artifacts from any previous build attempts
RUN if [ "$BUILDFLAG" != "azure" ]; then \
    cargo clean; \
    fi

# Clean previous build artifacts and build in one step
RUN if [ "$BUILDFLAG" != "azure" ]; then \
    cargo clean && \
    if [ "$TARGETARCH" = "arm" ]; then \
        CC=arm-linux-gnueabihf-gcc \
        CARGO_BUILD_TARGET=armv7-unknown-linux-gnueabihf \
        RUSTFLAGS="-C linker=arm-linux-gnueabihf-gcc" \
        cargo build --release --target=armv7-unknown-linux-gnueabihf && \
        cp target/armv7-unknown-linux-gnueabihf/release/*.so /extensions/; \
    elif [ "$TARGETARCH" = "arm64" ]; then \
        CC=aarch64-linux-gnu-gcc \
        CARGO_BUILD_TARGET=aarch64-unknown-linux-gnu \
        cargo build --release --target=aarch64-unknown-linux-gnu && \
        cp target/aarch64-unknown-linux-gnu/release/*.so /extensions/; \
    else \
        cargo build --release && \
        cp target/release/*.so /extensions/; \
    fi; \
    fi

WORKDIR /workspace/api

# Install gofail
RUN if [ "$FAULT_INJECTION_ENABLED" = "true" ]; then \
    go install go.etcd.io/gofail@latest && \
    find /workspace/api -type d | while read -r dir; do gofail enable $dir; done && \
    find /workspace/coa -type d | while read -r dir; do gofail enable $dir; done && \
    cd /workspace/api && go get go.etcd.io/gofail/runtime; \
fi

RUN if [ "$TARGETARCH" = "arm" ]; then \
    CC=arm-linux-gnueabihf-gcc \
    RUSTFLAGS="-C linker=arm-linux-gnueabihf-gcc" \
    CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} GODEBUG=netdns=cgo go build -tags "$BUILDFLAG" -o /dist/symphony-api; \
elif [ "$TARGETARCH" = "arm64" ]; then \
    CC=aarch64-linux-gnu-gcc \
    CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} GODEBUG=netdns=cgo go build -tags "$BUILDFLAG" -o /dist/symphony-api; \
else \
    CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} GODEBUG=netdns=cgo go build -tags "$BUILDFLAG" -o /dist/symphony-api; \
fi

FROM ${TARGET_BASE_IMAGE}

ARG TARGETARCH

RUN \
    set -x \
    && apt-get update \
    && apt-get install -y gcc g++ openssl libc6 libc6-dev libssl-dev ca-certificates curl wget jq android-tools-adb binutils \
    && if [ "${TARGETARCH}" = "arm64" ]; then \
        wget -O helm-v3.16.2-linux-arm64.tar.gz https://get.helm.sh/helm-v3.16.2-linux-arm64.tar.gz && \
        tar -zxvf helm-v3.16.2-linux-arm64.tar.gz && \
        mv linux-arm64/helm /usr/local/bin/helm && \
        chmod +x /usr/local/bin/helm; \
    else \
        curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 && \
        chmod +x get_helm.sh && \
        ./get_helm.sh; \
    fi \
    && apt upgrade -y
ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /zoneinfo.zip
ENV ZONEINFO=/zoneinfo.zip
COPY --from=build /dist /
COPY --from=build /extensions /extensions
RUN mkdir /remote-agent
COPY --from=build /remote-agent /v1alpha2/files
ADD ./api/symphony-api.json /
ADD ./api/symphony-api-no-k8s.json /
EXPOSE 8080
EXPOSE 8081
ENV LOG_LEVEL=Debug
ENV GOFAIL_HTTP="127.0.0.1:22381"
# ENV CONFIG /symphony-api.json
CMD sh -c 'if [ -f /etc/pki/ca-trust/source/anchors/proxy-cert.crt ]; then update-ca-trust; fi && exec /symphony-api -c $CONFIG -l $LOG_LEVEL'