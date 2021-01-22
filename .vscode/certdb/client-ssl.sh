echo -e 'GET /mutate-v1-pod HTTP/1.1\r\n\r\n' |  openssl s_client -CAfile ca.crt -connect ${HOSTNAME}:9443
