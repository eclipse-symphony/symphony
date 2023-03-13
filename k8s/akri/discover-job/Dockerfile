FROM mcr.microsoft.com/cbl-mariner/base/core:2.0
RUN \
    set -x \
    && apt-get update \
    && apt-get install -y openssl libssl-dev ca-certificates
ADD ./discover-job /

CMD ["/discover-job"]