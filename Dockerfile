FROM ubuntu:16.04
MAINTAINER Barna Csorogi <barnacs@justletit.be>

RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get upgrade -y && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        libjpeg8 \
        openssl \
        ssl-cert && \
    DEBIAN_FRONTEND=noninteractive apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /opt/compy
COPY \
    compy \
    docker.sh \
    /opt/compy/

# TODO: configure HTTP BASIC authentication
# TODO: user-provided certificates
ENV \
    CERTIFICATE_DOMAIN="localhost"

EXPOSE 9999
ENTRYPOINT ["./docker.sh"]
