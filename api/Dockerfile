##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

# Use Ubuntu as the base for the build stage to ensure proper cross-compilation support
ARG BUILD_BASE_IMAGE=ubuntu:22.04
ARG TARGET_BASE_IMAGE=ubuntu:latest

FROM --platform=$BUILDPLATFORM ${BUILD_BASE_IMAGE} AS build

ARG TARGETPLATFORM
ARG TARGETARCH

ENV CGO_ENABLED=1
ENV GO_VERSION=1.22.3
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

# Install necessary build tools, cross-compilers, and Go
RUN apt-get update && apt-get install -y \
    gcc g++ make curl wget git cargo \
    gcc-arm-linux-gnueabihf gcc-aarch64-linux-gnu \
    libssl-dev pkg-config \
    && wget https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

# Install Rust and rustup, overriding any existing Rust installation
RUN curl https://sh.rustup.rs -sSf | sh -s -- -y --no-modify-path \
    && . "$HOME/.cargo/env" \
    && rustup target add armv7-unknown-linux-gnueabihf aarch64-unknown-linux-gnu

# Set up Rust environment
ENV CARGO_TARGET_ARM_UNKNOWN_LINUX_GNUEABIHF_LINKER=arm-linux-gnueabihf-gcc
ENV CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc
ENV PATH=/root/.cargo/bin:$PATH

# Set the correct cross-compiler based on the target architecture
RUN if [ "$TARGETARCH" = "arm" ]; then \
        export CC=arm-linux-gnueabihf-gcc; \
        export CARGO_BUILD_TARGET=armv7-unknown-linux-gnueabihf; \
    elif [ "$TARGETARCH" = "arm64" ]; then \
        export CC=aarch64-linux-gnu-gcc; \
        export CARGO_BUILD_TARGET=aarch64-unknown-linux-gnu; \
    else \
        export CC=gcc; \
    fi

RUN mkdir /workspace
COPY ./packages /workspace/packages
COPY ./coa /workspace/coa
COPY ./api /workspace/api

# Build the Rust component
WORKDIR /workspace/api/pkg/apis/v1alpha1/providers/target/rust

# Clean build artifacts from any previous build attempts
RUN cargo clean

RUN if [ "$TARGETARCH" = "arm" ]; then \
        cargo build --release --target=armv7-unknown-linux-gnueabihf; \
    elif [ "$TARGETARCH" = "arm64" ]; then \
        cargo build --release --target=aarch64-unknown-linux-gnu; \
    else \
        cargo build --release; \
    fi

WORKDIR /workspace/api

# File permissions are not preserved when copying files in ADO.
RUN chmod +x pkg/apis/v1alpha1/providers/target/script/mock-*.sh

RUN if [ "$TARGETARCH" = "arm" ]; then \
        CC=arm-linux-gnueabihf-gcc; \
        export LIBRARY_PATH=/workspace/api/pkg/apis/v1alpha1/providers/target/rust/target/armv7-unknown-linux-gnueabihf/release:$LIBRARY_PATH; \
    elif [ "$TARGETARCH" = "arm64" ]; then \
        CC=aarch64-linux-gnu-gcc; \
        export LIBRARY_PATH=/workspace/api/pkg/apis/v1alpha1/providers/target/rust/target/aarch64-unknown-linux-gnu/release:$LIBRARY_PATH; \
    else \
        CC=gcc; \
        export LIBRARY_PATH=/workspace/api/pkg/apis/v1alpha1/providers/target/rust/target/release:$LIBRARY_PATH; \
    fi && \
    CGO_ENABLED=${CGO_ENABLED} GOOS=linux GOARCH=${TARGETARCH} GODEBUG=netdns=cgo CC=$CC go build -o /dist/symphony-api


FROM ${TARGET_BASE_IMAGE}

RUN set -x \
    && apt-get update \
    && apt-get install -y gcc g++ openssl libc6 libc6-dev libssl-dev ca-certificates curl jq android-tools-adb binutils \
    && curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 \
    && chmod +x get_helm.sh \
    && ./get_helm.sh \
    && apt upgrade -y

ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /zoneinfo.zip
ENV ZONEINFO /zoneinfo.zip
COPY --from=build /dist /
ADD ./api/symphony-api.json /
EXPOSE 8080
EXPOSE 8081
ENV LOG_LEVEL Debug

CMD sh -c 'if [ -f /etc/pki/ca-trust/source/anchors/proxy-cert.crt ]; then update-ca-trust; fi && exec /symphony-api -c $CONFIG -l $LOG_LEVEL'
