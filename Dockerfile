FROM ubuntu:16.04
MAINTAINER Barna Csorogi <barnacs@justletit.be>

RUN DEBIAN_FRONTEND=noninteractive apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get upgrade -y
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y \
    libjpeg8 \
    openssl \
    ssl-cert

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
