FROM ubuntu:16.04 as compy-builder
MAINTAINER Barna Csorogi <barnacs@justletit.be>

RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get upgrade -y && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        curl \
        g++ \
        git \
        libjpeg8-dev

RUN mkdir -p /usr/local/ && \
    curl -O https://storage.googleapis.com/golang/go1.9.linux-amd64.tar.gz && \
    tar xf go1.9.linux-amd64.tar.gz -C /usr/local

RUN mkdir -p /root/go/src/github.com/barnacs/compy/
COPY . /root/go/src/github.com/barnacs/compy/
WORKDIR /root/go/src/github.com/barnacs/compy
RUN /usr/local/go/bin/go get -d -v ./...
RUN /usr/local/go/bin/go build -v

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
    --from=compy-builder \
    /root/go/src/github.com/barnacs/compy/compy \
    /root/go/src/github.com/barnacs/compy/docker.sh \
    /opt/compy/

# TODO: configure HTTP BASIC authentication
# TODO: user-provided certificates
ENV \
    CERTIFICATE_DOMAIN="localhost"

EXPOSE 9999
ENTRYPOINT ["./docker.sh"]
