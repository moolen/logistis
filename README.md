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
- [x] `kubectl blame` plugin to inspect changes
  - [x] overview on a cluster-global level: latest changes by namespace
  - [x] latest changes per-namespace
* add `/lock` API to lock down namespaces or resources from particular users
* distributed event storage 🤷
- [x] make event limit configurable (currently max 100 events per resource)
- add predicate functions to decide whether or not to capture a event

### Quickstart

```
$ make deploy
# ...make changes to a pod or deployment
```

# blame with `patch`

```
kubectl blame deployment
+--------------------------+-----------+------------------+----------------------+---------------------------------------------------------------------+
| KEY                      | OPERATION | USER             | TIME                 | PATCH                                                               |
+--------------------------+-----------+------------------+----------------------+---------------------------------------------------------------------+
| default/Deployment/nginx | UPDATE    | kubernetes-admin | 2022-08-13T22:08:53Z | replace /metadata/labels/app                                        |
|                          |           |                  |                      | remove /metadata/managedFields/0/fieldsV1/f:metadata/f:labels/.     |
|                          |           |                  |                      | remove /metadata/managedFields/0/fieldsV1/f:metadata/f:labels/f:app |
|                          |           |                  |                      | add /metadata/managedFields/2                                       |
|                          |           |                  |                      |                                                                     |
+--------------------------+-----------+------------------+----------------------+---------------------------------------------------------------------+
| default/Deployment/nginx | CREATE    | kubernetes-admin | 2022-08-13T22:08:45Z | add                                                                 |
|                          |           |                  |                      |                                                                     |
+--------------------------+-----------+------------------+----------------------+---------------------------------------------------------------------+
```

# blame with `diff`

```
kubectl blame deployment -f diff
+---------------------------------------------------------------------------------+
| 35m | default/deployment/nginx | UPDATE | kubernetes-admin                      |
| groups:                                                                         |
| - system:masters                                                                |
| - system:authenticated                                                          |
|                                                                                 |
+---------------------------------------------------------------------------------+
|        "deployment.kubernetes.io/revision": "1",                                |
|        "kubectl.kubernetes.io/last-applied-configuration": "{"apiVersion":"apps |
| "                                                                               |
|      },                                                                         |
|      "creationTimestamp": "2022-09-15T20:09:52Z",                               |
| -    "generation": 5,                                                           |
| +    "generation": 6,                                                           |
|      "managedFields": [                                                         |
|        0: {                                                                     |
|          "apiVersion": "apps/v1",                                               |
|          "fieldsType": "FieldsV1",                                              |
|          "fieldsV1": {                                                          |
| [...]                                                                           |
|              }                                                                  |
|            },                                                                   |
|            "f:spec": {                                                          |
|              "f:progressDeadlineSeconds": {                                     |
|              },                                                                 |
| -            "f:replicas": {                                                    |
| -            },                                                                 |
|              "f:revisionHistoryLimit": {                                        |
|              },                                                                 |
|              "f:selector": {                                                    |
|              },                                                                 |
|              "f:strategy": {                                                    |
| [...]                                                                           |
|          "manager": "kube-controller-manager",                                  |
|          "operation": "Update",                                                 |
|          "subresource": "status",                                               |
|          "time": "2022-09-17T19:37:54Z"                                         |
|        }                                                                        |
| +      2: {                                                                     |
| +        "apiVersion": "apps/v1",                                               |
| +        "fieldsType": "FieldsV1",                                              |
| +        "fieldsV1": {                                                          |
| +          "f:spec": {                                                          |
| +            "f:replicas": {                                                    |
| +            }                                                                  |
| +          }                                                                    |
| +        },                                                                     |
| +        "manager": "kubectl-edit",                                             |
| +        "operation": "Update",                                                 |
| +        "time": "2022-09-17T19:37:59Z"                                         |
| +      }                                                                        |
|      ],                                                                         |
|      "name": "nginx",                                                           |
|      "namespace": "default",                                                    |
|      "resourceVersion": "178315",                                               |
|      "uid": "4dc72be3-0a50-4fdb-b2b6-b9f50fb60976"                              |
|    },                                                                           |
|    "spec": {                                                                    |
|      "progressDeadlineSeconds": 600,                                            |
| -    "replicas": 2,                                                             |
| +    "replicas": 1,                                                             |
|      "revisionHistoryLimit": 10,                                                |
|      "selector": {                                                              |
|        "matchLabels": {                                                         |
|          "app": "nginx"                                                         |
|        }                                                                        |
|                                                                                 |
+---------------------------------------------------------------------------------+
```

### Custom Base Image

Build with your own custom image with a custom Makefile:

```Makefile
# .local/Makefile

IMAGE_REPO := custom.acme.org/logistis
IMAGE_TAG := v0.0.0-dev

DOCKER_BUILD_ARGS := --load \
	--build-arg BASEIMAGE=custom.acme.org/golang:xyz \
	--build-arg RUNIMAGE=custom.acme.org/alpine:xyz

export

.DEFAULT:
	make -C ../ $@
```