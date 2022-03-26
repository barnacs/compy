FROM golang:1.12-alpine as builder

WORKDIR /root/go/src/github.com/barnacs/compy/

COPY . .

RUN apk add --no-cache --no-progress git g++ libjpeg-turbo-dev
RUN go get -d -v ./...
RUN go build -ldflags='-extldflags "-static" -s -w' -o /go/bin/compy


FROM scratch

LABEL maintainer="Sandro Jäckel <sandro.jaeckel@gmail.com>"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/compy compy

ENTRYPOINT ["./compy"]
