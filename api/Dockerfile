##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

ARG BUILD_BASE_IMAGE=golang:1.22.3-alpine
ARG TARGET_BASE_IMAGE=ubuntu:latest

FROM --platform=$BUILDPLATFORM ${BUILD_BASE_IMAGE} AS build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=1

# Install gcc, musl-dev, and the appropriate cross-compiler based on architecture
RUN apk add --no-cache gcc musl-dev g++ cargo \
    && if [ "$TARGETARCH" = "arm" ]; then \
           apk add --no-cache binutils-arm-none-eabi gcc-arm-none-eabi; \
           export CC=arm-none-eabi-gcc; \
       elif [ "$TARGETARCH" = "arm64" ]; then \
           apk add --no-cache gcc-aarch64-linux-gnu; \
           export CC=aarch64-linux-gnu-gcc; \
       else \
           export CC=gcc; \
       fi

RUN mkdir /workspace
COPY ./packages /workspace/packages
COPY ./coa /workspace/coa
COPY ./api /workspace/api

# Build the Rust component
WORKDIR /workspace/api/pkg/apis/v1alpha1/providers/target/rust
RUN cargo build --release

WORKDIR /workspace/api

# File permissions are not preserved when copying files in ADO. 
RUN chmod +x pkg/apis/v1alpha1/providers/target/script/mock-*.sh
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} GODEBUG=netdns=cgo go build -o /dist/symphony-api

FROM ${TARGET_BASE_IMAGE}
RUN \
    set -x \
    && apt-get update \
    && apt-get install -y openssl libssl-dev ca-certificates curl jq  android-tools-adb \
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
# ENV CONFIG /symphony-api.json
CMD sh -c 'if [ -f /etc/pki/ca-trust/source/anchors/proxy-cert.crt ]; then update-ca-trust; fi && exec /symphony-api -c $CONFIG -l $LOG_LEVEL'
