FROM scratch

COPY apollo /

EXPOSE 8080

ENTRYPOINT ["/apollo"]
