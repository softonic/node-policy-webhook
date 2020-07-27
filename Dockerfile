FROM golang:1.14-buster AS build

ENV GOBIN=$GOPATH/bin

ADD . /src/node-policy-webhook

WORKDIR /src/node-policy-webhook

RUN make build

FROM debian:buster-slim

COPY --from=build /src/node-policy-webhook/node-policy-webhook /node-policy-webhook

ENTRYPOINT ["/node-policy-webhook"]
