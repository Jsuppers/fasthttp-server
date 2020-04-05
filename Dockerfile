ARG PROJECT=fasthttp-server

FROM golang@sha256:cee6f4b901543e8e3f20da3a4f7caac6ea643fd5a46201c3c2387183a332d989 as builder

ARG PROJECT
WORKDIR /go/${PROJECT}
COPY . /go/${PROJECT}
RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates && rm -rf /var/cache/apk/* && apk add make && apk add git
RUN make build

FROM scratch

ARG PROJECT
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/${PROJECT}/bin/${PROJECT} service

CMD ["./service"]

