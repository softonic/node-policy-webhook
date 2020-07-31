Webhook Makefile will help you build and deploy in your local cluster.
Deployment in production could be done with helm public chart

#### Build

```sh
$ make dev
``` 


Go mod download to get dependancies
Go build 
Build image with just final go binary using multistages.
Use the image generated before and load the image into kind local registry

#### Deploy 

```sh
$ make deploy
```

Use kubebuilder controller-gen to generate CRD
Generate secrets. The script first retrieve the CA from apiserver, and then generate csr and key. Request api to sign and generate the public certificate we will use in our webhook https server.
Generate secrets values for helm.
With helm template, generate a file called manifest.yaml
Apply with kubectl this manifest.yaml

You can alternatively deploy with helm, with 

```sh
$ make helm-deploy
```
