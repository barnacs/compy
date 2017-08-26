#!/bin/sh

openssl req -x509 -newkey rsa:2048 -nodes -keyout cert.key -out cert.crt -days 3650 -subj "/CN=${CERTIFICATE_DOMAIN}"
openssl req -x509 -newkey rsa:2048 -nodes -keyout ca.key -out ca.crt -days 3650 -subj "/CN=${CERTIFICATE_DOMAIN}"

echo 'Generated server certificate:'
cat cert.crt
echo
echo 'Generated CA certificate:'
cat ca.crt

exec ./compy \
    -cert cert.crt -key cert.key \
    -ca ca.crt -cakey ca.key \
    :9999
