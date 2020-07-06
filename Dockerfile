FROM debian:buster

ADD bin/linux_amd64/node-policy-webhook /node-policy-webhook

USER 1000

ENTRYPOINT ["/node-policy-webhook"]
