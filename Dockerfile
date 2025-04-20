FROM golang:1.23

WORKDIR ${GOPATH}/avito-pvz/
COPY . ${GOPATH}/avito-pvz/

RUN go build -o /build ./cmd  && \
go clean -cache -modcache

EXPOSE 8080 9000 3000

CMD ["/build"]
