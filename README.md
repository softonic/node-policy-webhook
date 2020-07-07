# node-policy-webhook
K8s webhook handling profiles for tolerations, nodeSelector and nodeAffinity


## TEST

```bash

kubectl label namespace default mutateme-
make cert
make apply-path
make install
make push  ### docker push to your favourite registry
make deploy
kubectl label namespace default mutateme=enabled
```

then you can test if the mutating works

```bash
kubectl run --generator=run-pod/v1 -it --rm=true kdgb2 --restart=Never --image=nvucinic/kdbg --  /bin/bash
```
