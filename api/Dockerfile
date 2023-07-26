## This is the Dockerfile for the symphony-api. It's intended build context is the root of the repository.
## Either build the image from the root of the repository:
##   docker build -f api/Dockerfile .
## Or build with docker-compose file of api.
FROM mcr.microsoft.com/devcontainers/go:1.20 AS build

RUN go install github.com/magefile/mage@latest
RUN mkdir /workspace
COPY ./packages /workspace/packages
COPY ./coa /workspace/coa
COPY ./api /workspace/api
WORKDIR /workspace/api
# File permissions are not preserved when copying files in ADO. 
RUN chmod +x pkg/apis/v1alpha1/providers/target/script/mock-*.sh

ENV GOOS=linux 
ENV GOARCH=amd64 
ARG DANGEROUSLY_SKIP_TESTS=false

RUN [ "$DANGEROUSLY_SKIP_TESTS" = "true" ] || CGO_ENABLED=1 mage TestWithCoa
RUN CGO_ENABLED=0 mage build

FROM mcr.microsoft.com/cbl-mariner/base/core:2.0
RUN \
    set -x \
    && tdnf check-update \
    && tdnf install -y ca-certificates openssl curl jq awk tar 

WORKDIR /
COPY --from=build /workspace/api/bin/symphony-api .
EXPOSE 8080
EXPOSE 8081
ENV LOG_LEVEL Debug
CMD exec /symphony-api -c $CONFIG -l $LOG_LEVEL