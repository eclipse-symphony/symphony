FROM --platform=$BUILDPLATFORM mcr.microsoft.com/devcontainers/go AS build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0

RUN mkdir /src
WORKDIR /src
COPY . .
 
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} GODEBUG=netdns=cgo go build -o /dist/symphony-api

FROM mcr.microsoft.com/cbl-mariner/base/core:2.0
RUN \
    set -x \
    && tdnf check-update \
    && tdnf install -y ca-certificates openssl libssl-dev curl jq  android-tools-adb \
    && curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 \
    && chmod +x get_helm.sh \
    && ./get_helm.sh \
    && tdnf upgrade -y

COPY --from=build /dist /
ADD ./symphony-api.json /
EXPOSE 8080
EXPOSE 8081
ENV LOG_LEVEL Debug
ENV CONFIG /symphony-api.json
CMD exec /symphony-api -c $CONFIG -l $LOG_LEVEL