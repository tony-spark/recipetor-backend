FROM golang:1.19-alpine3.17 AS builder

WORKDIR /usr/local/go/src/

ADD . /usr/local/go/src/

RUN go clean --modcache
RUN go build -mod=readonly -o nutrition-facts-srv cmd/nutrition-facts-srv/main.go

FROM alpine:3.17

COPY --from=builder /usr/local/go/src/nutrition-facts-srv /

CMD ["/nutrition-facts-srv"]