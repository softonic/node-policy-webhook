FROM golang:1.15-buster AS build

ENV GOBIN=$GOPATH/bin

ADD . /src/k8s-policy-controller

WORKDIR /src/k8s-policy-controller

RUN make build

FROM debian:buster-slim

COPY --from=build /src/k8s-policy-controller/k8s-policy-controller /k8s-policy-controller

ENTRYPOINT ["/k8s-policy-controller"]
