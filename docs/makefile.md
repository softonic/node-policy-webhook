Webhook Makefile will help you build and deploy in your local cluster.
Deployment in production could be done with helm public chart

#### Build

```sh
$ make dev
``` 

Dev Steps:
* Go mod download to get dependancies
* Go build. Binary will be placed in the cwd
* Build image with just final go binary using multistages.
* Use the image generated before and load the image into kind local registry

#### Deployment with kubectl

Modify the Makefile to point to the public image repository. Set the tag also via VERSION variable.

```bash
VERSION ?= 0.1.1
REPOSITORY ?= softonic/node-policy-webhook
```

```sh
$ make deploy
```

You can also use your local registry, loading then to kind. Or private registry using docker push. 

Deploy steps:
* Use kubebuilder controller-gen to generate CRD
* Generate secrets. The script first retrieve the CA from apiserver, and then generate csr and key. Request api to sign and generate the public certificate we will use in our webhook https server.
* Generate secrets values for helm.
* With helm template, generate a file called manifest.yaml
* Apply with kubectl this manifest.yaml

You can alternatively deploy with helm, with 

```sh
$ make helm-deploy
```

#### Deployment with helm public chart

```bash
$ helm repo add softonic https://charts.softonic.io
```

You should generate your secrets, that will be composed of a certificate TLS and a private key. This data is not in the helm public chart. 
Run the generate secrets script and you will find the files in the ssl directory.
Then just run this command to generate the tls secret.

```bash
$ kubectl create ns node-policy-webhook
$ kubectl create secret generic node-policy-webhook --from-file=key.pem=ssl/node-policy-webhook.key --from-file=cert.pem=ssl/node-policy-webhook.pem --dry-run -o yaml | kubectl -n node-policy-webhook apply -f -
```


```bash
$ helm install --namespace node-policy-webhook --name my-release softonic/node-policy-webhook
```


