FROM golang:1.23-alpine AS go-build
# FROM golang:1.15 AS go-build

WORKDIR /go/src/github.com/hunterhug/fafacms

ADD . /go/src/github.com/hunterhug/fafacms

RUN GOOS=linux GOARCH=amd64 go build -v -ldflags "-s -w" -o fafacms main.go

FROM alpine:3.20 AS prod
# not worth use gcc image, but i love it
# https://www.debian.org/releases/
# FROM bitnami/minideb:buster
#FROM bitnami/minideb-extras-base:stretch-r165 AS prod

WORKDIR /root/

COPY --from=go-build /go/src/github.com/hunterhug/fafacms/fafacms /bin/fafacms

RUN chmod 777 /bin/fafacms

CMD /bin/fafacms $RUN_OPTS
