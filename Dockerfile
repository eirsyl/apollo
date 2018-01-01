FROM scratch

COPY bin/apollo /

ENTRYPOINT ["/apollo"]
