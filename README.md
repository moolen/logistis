# 📖 Logistis

User Interaction Tracking for Kubernetes.

Logistis is an application which observes **user interaction** with a Kubernetes cluster. Think `Google Analytics for Kubernetes`.
It tracks CRUD operations via ValidatingWebhook and stores it in its own local database.
The tracking events can be fetched as `json` or shown as `diff`.

It allows you to:
* track who made a change in which namespace to which resource
* track what changes have been made and when
* get an overview of recent changes in a cluster

> Yes, everyone does gitops. But people still need to operate their stuff.

#### TODO/Roadmap
* `kubectl blame` plugin to inspect changes
  * overview on a cluster-global level: latest changes by namespace
  * latest changes per-namespace
* add `/lock` API to lock down namespaces or resources from particular users
* distributed event storage 🤷
* make event limit configurable (currently max 100 events per resource)

### Quickstart

```
$ make deploy
# ...make changes to a pod or deployment
```

#### fetch events
```
$ curl -k "https://localhost:8888/events" | jq
{
  "apps//Pod/lifespan-seven": [
    {
      "id": "0f5a2222-69a2-43d5-9dbe-85dbc23491fa",
      "group": "",
      "kind": "Pod",
      "namespace": "apps",
      "name": "lifespan-seven",
      "time": "2022-08-12T23:36:30.118019949Z",
      "operation": "UPDATE",
      "userInfo": "&UserInfo{Username:kubernetes-admin,UID:,Groups:[system:masters system:authenticated],Extra:map[string]ExtraValue{},}",
      "object": "...",
      "oldObject": "..."
```

#### show diff
```diff
$ k port-forward svc/simple-kubernetes-webhook 8888:443 &
$ curl -ik "https://localhost:8888/diff?namespace=kube-system"
HTTP/2 200
content-type: text/plain; charset=utf-8
content-length: 828
date: Fri, 12 Aug 2022 23:57:02 GMT

\\\\\\\\\\\\\\\\\\\\
kube-system/apps/Deployment/coredns
\\\\\\\\\\\\\\\\\\\\
operation: UPDATE
time: 2022-08-12 23:41:54.265330053 +0000 UTC
userinfo: &UserInfo{Username:kubernetes-admin,UID:,Groups:[system:masters system:authenticated],Extra:map[string]ExtraValue{},}
     "labels": {
       "foo": "13123123",
       "k8s-app": "kube-dns"
+      "fart": "12312313"
     },
     "managedFields": [
       0: {
[...]
             "f:labels": {
               "f:foo": {
               }
+              "f:fart": {
+              }
             }
           }
         },
         "manager": "kubectl-edit",
         "operation": "Update",
-        "time": "2022-08-12T23:41:19Z"
+        "time": "2022-08-12T23:41:54Z"
       }
     ],
     "name": "coredns",

---
```