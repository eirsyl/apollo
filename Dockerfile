FROM scratch

ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY bin/apollo /

ENTRYPOINT ["/apollo"]
