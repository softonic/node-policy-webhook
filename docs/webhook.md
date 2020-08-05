
### Our webhook

What is a mutating webhook and how does it work?

Webhook, or mutating webhook in our case, is composed of 2 parts.
One is the definition of the mutating webhook. 
MutatingWebhookConfiguration describes the configuration of and admission webhook. And with this configuration we tell
k8s that we want to receive request when a pod is about to be created, and to tell api to send the request ( Admission Review ) to the /mutate url at 443 port.

```
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
...
webhooks:
...
    clientConfig:
      caBundle: ${CA_BUNDLE}
      service:
        name: node-policy-webhook
        namespace: node-policy-webhook
        path: "/mutate"
        port: 443
    rules:
      - operations: ["CREATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
...
```

Another part is the service that receives the requests and do the job. In our case, the job is to "mutate" the pods that 
are being scheduled in the system. Everytime a pod is ready to be scheduled, our service receives a request from the 
k8s api server. Our main work here, is to tell the api if we want to modify this object.

What does we would want to modify?

Specifically, we would like to modify parts of the object that are related with assigning pods to nodes.
These are, 

* nodeSelectors
* Tolerations
* NodeAffinities


What or when we will take place the modification of these PodSpec's fields ? 

webhook will check if the Pod has the annotation

```
"nodepolicy.softonic.io/profile"
```

If the annotation does not exist, pod will be schedule with no mutation.
If the annotation does exist, it will get the value of the annotation. 
With this value we can match the NodePolicyProfile ( our CRD ).
Webhook will patch the nodeSelector, NodeAffiniy and tolerations of the NodePolicyProfile
that we are setting with the profile annotation.

More info:

https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/
