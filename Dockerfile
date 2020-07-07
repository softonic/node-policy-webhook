FROM debian:buster

ADD bin/linux_amd64/node-policy-webhook /node-policy-webhook


EXPOSE 8080

ENTRYPOINT ["/node-policy-webhook"]
