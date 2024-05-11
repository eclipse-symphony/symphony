##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

## The build context is expected to be the root of the repository.
## Either build the image from the root of the repository:
##   docker build -f k8s/Dockerfile .
## Or build with docker-compose file of k8s

ARG BUILD_BASE_IMAGE=golang:1.20.2-alpine
ARG TARGET_BASE_IMAGE=mcr.microsoft.com/mirror/docker/library/alpine:3.16

FROM --platform=$BUILDPLATFORM ${BUILD_BASE_IMAGE} AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0

# Install gcc, g++ and other necessary build tools
RUN apk add --no-cache gcc musl-dev

RUN go install github.com/magefile/mage@latest
WORKDIR /
# Copy the Go Modules 
COPY api/ api/
COPY coa/ coa/
COPY packages/ packages/
COPY k8s/ k8s/
WORKDIR /k8s
# Test
RUN CGO_ENABLED=1 mage generate operatorTest
# Build
RUN CGO_ENABLED=0 mage build
FROM ${TARGET_BASE_IMAGE} AS manager
WORKDIR /
COPY --from=builder /k8s/bin/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
