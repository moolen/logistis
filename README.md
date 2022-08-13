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
kubectl blame --target-namespace=default
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
kubectl blame --target-namespace=default --format=diff
+--------------------------+-----------+------------------+----------------------+----------------------------------------------------------------+
| KEY                      | OPERATION | USER             | TIME                 | DIFF                                                           |
+--------------------------+-----------+------------------+----------------------+----------------------------------------------------------------+
| default/Deployment/nginx | UPDATE    | kubernetes-admin | 2022-08-13T22:08:53Z |        "deployment.kubernetes.io/revision": "1"                |
|                          |           |                  |                      |      },                                                        |
|                          |           |                  |                      |      "creationTimestamp": "2022-08-13T22:08:45Z",              |
|                          |           |                  |                      |      "generation": 1,                                          |
|                          |           |                  |                      |      "labels": {                                               |
|                          |           |                  |                      | -      "app": "nginx"                                          |
|                          |           |                  |                      | +      "app": "nginx123"                                       |
|                          |           |                  |                      |      },                                                        |
|                          |           |                  |                      |      "managedFields": [                                        |
|                          |           |                  |                      |        0: {                                                    |
|                          |           |                  |                      |          "apiVersion": "apps/v1",                              |
|                          |           |                  |                      |          "fieldsType": "FieldsV1",                             |
|                          |           |                  |                      |          "fieldsV1": {                                         |
|                          |           |                  |                      |            "f:metadata": {                                     |
|                          |           |                  |                      |              "f:labels": {                                     |
|                          |           |                  |                      | -              ".": {                                          |
|                          |           |                  |                      | -              },                                              |
|                          |           |                  |                      | -              "f:app": {                                      |
|                          |           |                  |                      | -              }                                               |
|                          |           |                  |                      |              }                                                 |
|                          |           |                  |                      |            },                                                  |
|                          |           |                  |                      |            "f:spec": {                                         |
|                          |           |                  |                      |              "f:progressDeadlineSeconds": {                    |
|                          |           |                  |                      |              },                                                |
|                          |           |                  |                      | [...]                                                          |
|                          |           |                  |                      |          "manager": "kube-controller-manager",                 |
|                          |           |                  |                      |          "operation": "Update",                                |
|                          |           |                  |                      |          "subresource": "status",                              |
|                          |           |                  |                      |          "time": "2022-08-13T22:08:47Z"                        |
|                          |           |                  |                      |        }                                                       |
|                          |           |                  |                      | +      2: {                                                    |
|                          |           |                  |                      | +        "apiVersion": "apps/v1",                              |
|                          |           |                  |                      | +        "fieldsType": "FieldsV1",                             |
|                          |           |                  |                      | +        "fieldsV1": {                                         |
|                          |           |                  |                      | +          "f:metadata": {                                     |
|                          |           |                  |                      | +            "f:labels": {                                     |
|                          |           |                  |                      | +              "f:app": {                                      |
|                          |           |                  |                      | +              }                                               |
|                          |           |                  |                      | +            }                                                 |
|                          |           |                  |                      | +          }                                                   |
|                          |           |                  |                      | +        },                                                    |
|                          |           |                  |                      | +        "manager": "kubectl-edit",                            |
|                          |           |                  |                      | +        "operation": "Update",                                |
|                          |           |                  |                      | +        "time": "2022-08-13T22:08:53Z"                        |
|                          |           |                  |                      | +      }                                                       |
|                          |           |                  |                      |      ],                                                        |
|                          |           |                  |                      |      "name": "nginx",                                          |
|                          |           |                  |                      |      "namespace": "default",                                   |
|                          |           |                  |                      |      "resourceVersion": "643643",                              |
|                          |           |                  |                      |      "uid": "610b9e52-6612-417b-9dfc-4d4b1344aa61"             |
|                          |           |                  |                      |                                                                |
+--------------------------+-----------+------------------+----------------------+----------------------------------------------------------------+
```