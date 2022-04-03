FROM golang:alpine as build

RUN apk --update add ca-certificates
WORKDIR /go/src/kv-web
COPY . .
RUN CGO_ENABLED=0 go build -o kv .

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/src/kv-web/kv /kv

VOLUME ["/letsencrypt"]

EXPOSE 443/tcp

ENTRYPOINT ["/kv"]
