##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

ARG BUILD_BASE_IMAGE=golang:1.22.4-alpine
ARG TARGET_BASE_IMAGE=ubuntu:latest

FROM --platform=$BUILDPLATFORM ${BUILD_BASE_IMAGE} AS build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG FAULT_INJECTION_ENABLED=false

ENV CGO_ENABLED=0

RUN mkdir /workspace
COPY ./packages /workspace/packages
COPY ./coa /workspace/coa
COPY ./api /workspace/api
WORKDIR /workspace/api
# File permissions are not preserved when copying files in ADO. 
RUN chmod +x pkg/apis/v1alpha1/providers/target/script/mock-*.sh

# Install gofail
RUN if [ "$FAULT_INJECTION_ENABLED" == "true" ]; then \
    go install go.etcd.io/gofail@latest && \
    find /workspace/api -type d | while read -r dir; do gofail enable $dir; done && \
    find /workspace/coa -type d | while read -r dir; do gofail enable $dir; done && \
    cd /workspace/api && go get go.etcd.io/gofail/runtime; \
fi
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} GODEBUG=netdns=cgo go build -o /dist/symphony-api

FROM ${TARGET_BASE_IMAGE}

ARG TARGETARCH

RUN \
    set -x \
    && apt-get update \
    && apt-get install -y openssl libssl-dev ca-certificates curl jq wget android-tools-adb \
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
ADD ./api/symphony-api.json /
EXPOSE 8080
EXPOSE 8081
ENV LOG_LEVEL=Debug
ENV GOFAIL_HTTP="127.0.0.1:22381"
# ENV CONFIG /symphony-api.json
CMD sh -c 'if [ -f /etc/pki/ca-trust/source/anchors/proxy-cert.crt ]; then update-ca-trust; fi && exec /symphony-api -c $CONFIG -l $LOG_LEVEL'