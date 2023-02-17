FROM ubuntu:latest
RUN \
    set -x \
    && apt-get update \
    && apt-get install -y openssl libssl-dev ca-certificates
ADD ./discover-job /

CMD ["/discover-job"]