FROM golang:1.15-buster AS build

ENV GOBIN=$GOPATH/bin

ADD . /src/admission-webhook-controller

WORKDIR /src/admission-webhook-controller

RUN make build

FROM debian:buster-slim

COPY --from=build /src/admission-webhook-controller/admission-webhook-controller /admission-webhook-controller

ENTRYPOINT ["/admission-webhook-controller"]
