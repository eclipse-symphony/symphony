##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##
FROM --platform=$BUILDPLATFORM golang:1.20.2-alpine AS build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0

RUN mkdir /workspace
COPY ./packages /workspace/packages
COPY ./coa /workspace/coa
COPY ./api /workspace/api
WORKDIR /workspace/api
# File permissions are not preserved when copying files in ADO. 
RUN chmod +x pkg/apis/v1alpha1/providers/target/script/mock-*.sh
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} GODEBUG=netdns=cgo go build -o /dist/symphony-api

FROM ubuntu:latest
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
CMD exec /symphony-api -c $CONFIG -l $LOG_LEVEL


# ADO Docker file
# ## This is the Dockerfile for the symphony-api. It's intended build context is the root of the repository.
# ## Either build the image from the root of the repository:
# ##   docker build -f api/Dockerfile .
# ## Or build with docker-compose file of api.
# FROM mcr.microsoft.com/oss/go/microsoft/golang:1.21-fips-cbl-mariner2.0 AS build

# RUN go install github.com/magefile/mage@latest
# RUN tdnf install jq -y

# RUN mkdir /workspace
# COPY ./packages /workspace/packages
# COPY ./coa /workspace/coa
# COPY ./api /workspace/api
# WORKDIR /workspace/api
# # File permissions are not preserved when copying files in ADO. 
# RUN chmod +x pkg/apis/v1alpha1/providers/target/script/mock-*.sh

# # Setting environment configuration.
# ENV GOOS=linux 
# ENV GOARCH=amd64 
# ENV GOEXPERIMENT=systemcrypto
# ENV CGO_ENABLED=1
# ARG DANGEROUSLY_SKIP_TESTS=false

# # Test and build.
# RUN [ "$DANGEROUSLY_SKIP_TESTS" = "true" ] || mage unitTest
# RUN mage build

# FROM mcr.microsoft.com/cbl-mariner/base/core:2.0
# RUN \
#     set -x \
#     && tdnf check-update \
#     && tdnf install -y ca-certificates openssl curl jq awk tar 

# WORKDIR /
# COPY --from=build /workspace/api/bin/symphony-api .
# EXPOSE 8080
# EXPOSE 8081
# ENV LOG_LEVEL Debug
# CMD sh -c 'if [ -f /etc/pki/ca-trust/source/anchors/proxy-cert.crt ]; then update-ca-trust; fi && exec /symphony-api -c $CONFIG -l $LOG_LEVEL'
