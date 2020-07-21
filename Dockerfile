FROM golang:1.14-buster AS build

ADD . /src/node-policy-webhook

WORKDIR /src/node-policy-webhook

RUN go mod download &&\
 GOARCH=${ARCH} go install -ldflags "-X ${PKG}/pkg/version.Version=${VERSION}" ./cmd/node-policy-webhook/.../

FROM debian:buster-slim

COPY --from=build /src/node-policy-webhook/bin/linux_amd64/node-policy-webhook /node-policy-webhook

EXPOSE 443

USER 1000

ENTRYPOINT ["/node-policy-webhook"]
