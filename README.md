# node-policy-webhook
K8s webhook handling profiles for tolerations, nodeSelector and nodeAffinity


## DEVEL ENVIRONMENT

### Requirements

Install kind

```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.8.1/kind-$\(uname\)-amd64
mv kind-darwin-amd64 /usr/local/bin/kind
```


```bash

make dev
make deploy-dev
```

Now you can run a pod to test it
