FROM golang:1.19-alpine3.17 AS builder

WORKDIR /usr/local/go/src/

ADD . /usr/local/go/src/

RUN go clean --modcache
RUN go build -mod=readonly -o user-srv cmd/user-srv/main.go

FROM alpine:3.17

COPY --from=builder /usr/local/go/src/user-srv /

CMD ["/user-srv"]