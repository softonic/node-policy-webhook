[![Go Report Card](https://goreportcard.com/badge/softonic/node-policy-webhook)](https://goreportcard.com/report/softonic/node-policy-webhook)
[![Releases](https://img.shields.io/github/release-pre/softonic/node-policy-webhook.svg?sort=semver)](https://github.com/softonic/node-policy-webhook/releases)
[![LICENSE](https://img.shields.io/github/license/softonic/node-policy-webhook.svg)](https://github.com/softonic/node-policy-webhook/blob/master/LICENSE)
[![DockerHub](https://img.shields.io/docker/pulls/softonic/node-policy-webhook.svg)](https://hub.docker.com/r/softonic/node-policy-webhook)


# node-policy-webhook
K8s webhook handling profiles for tolerations, nodeSelector and nodeAffinity


## DEVEL ENVIRONMENT

### Requirements

Install kind

```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.8.1/kind-$\(uname\)-amd64
mv kind-darwin-amd64 /usr/local/bin/kind
kind create cluster
```

Compile the code and deploy the needed resources

```bash
make dev
make deploy
```

Now you can run a pod to test it


# Motivation


The goal of Node Policy Webhook is to reduce Kubernetes manifests complexity by 
moving the logic behind the scheduling ( when assigning pods to nodes ) 
to a unique place where the cluster's operator/admin can handle it. 

This is accomplished with a new mutating webhook and new CRD where you can place all the node scheduling intelligente 
in the form of profiles.

In this new CRD, you can set different "profiles" depending on the nodes or VMs provision on your cluster.

When doing that, you can remove all the tolterations,nodeSelectors and nodeAffinities from your 
manifests' workloads ( like Deployments, DaemonSets, StatefulSets ..)

If you are running Knative in your cluster, this project can help you. At the time this is being writen, 
there is no way you can schedule knative workloads ( in the form of ksvc ) to a desired node or vm 
as it does not implement tolerations neither nodeSelectors.


Example:

You hace an specific deployment, but you'd like these pods to be scheduled in nodes  with label disk=ssd

So first step is to create an object of type 

```
apiVersion: noodepolicies.softonic.io/v1alpha1
kind: NodePolicyProfile
metadata:
  name: ssd
spec:
  nodeSelector:
    disk: "ssd"
```


Now you just need to deploy these pods setting this annotation in your deployment

```
nodepolicy.softonic.io/profile: "ssd"
```

In deployment time, the mutating webhook will replace the nodeSelector with the nodeSelector above mentioned.


### Caveats


* If your workload already has a nodeSelector defined, mutating webhook will remove it.
* If your workload already has tolerations defined, mutating webhook will keep them.
* If your workload already has Affinities defined, it will keep the podAntiAffinities and podAffinities 
and will remove the nodeAffinities 





Assigning pods to nodes:

> You can constrain a Pod to only be able to run on particular Node(s), or to prefer to run on particular nodes. 
> There are several ways to do this, and the recommended approaches all use label selectors to make the selection. 
> Generally such constraints are unnecessary, as the scheduler will automatically do a reasonable 
> placement (e.g. spread your pods across nodes, not place the pod on a node with insufficient free resources, etc.) 
> but there are some circumstances where you may want more control on a node where a pod lands, for example to ensure 
> that a pod ends up on a machine with an SSD attached to it, or to co-locate pods from two different 
> services that communicate a lot into the same availability zone.



# Internals


* [Dynamic Webhooks](docs/internals.md)
* [Our Webhook](docs/webhook.md)
* [Makefile - Build and Deploy](docs/makefile.md)
