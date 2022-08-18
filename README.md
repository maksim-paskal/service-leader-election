# Sidecar to switch trafic to leader pod

Inspired by article [Simple leader election with Kubernetes and Docker](https://kubernetes.io/blog/2016/01/simple-leader-election-with-kubernetes/)

This tool can switch traffic to leader pod, for example if your application can not be running in multiple pods, but you want to make it more reliable, you can run your application in multiple replicas, but traffic will be process only by one of this pod.

## How it works

for example you have 3 replicas of your application, and you have a service, by default service will be proxy all traffic to all of your pods

```bash
NAME                                           READY   STATUS    RESTARTS   AGE
pod/service-leader-election-689f7875f6-fgmjb   1/1     Running   0          9m10s
pod/service-leader-election-689f7875f6-nsdcg   1/1     Running   0          9m10s
pod/service-leader-election-689f7875f6-zp4wp   1/1     Running   0          9m10s

NAME                              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
service/service-leader-election   ClusterIP   10.100.54.205   <none>        28086/TCP   9m11s
```

this sidecar:

1) label your pod with additional label `service-leader-election=<pod name>`
2) elects leader between your pods
3) switch your service to leader pod

after election all traffic will be processed only by one of your pods

## Installation

you need additional serviceaccount that can get pod and service information and also can make a patch

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-leader-election
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: service-leader-election-role
rules:
- apiGroups: [""]
  resources: ["pods","services"]
  verbs: ["get", "patch"]
- apiGroups: ["coordination.k8s.io"]
  resources: ["leases"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: service-leader-election
roleRef:
  kind: Role
  name: service-leader-election-role
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  name: service-leader-election
  namespace: {{ .Release.Namespace }}
```

add additional sidecar to your deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: app
spec:
...
    spec:
      serviceAccountName: service-leader-election
      containers:
      - name: service-leader-election
        image: paskalmaksim/service-leader-election:latest
        imagePullPolicy: Always
        args:
        args:
        - -lease-name=lockname
        - -service-name=servicename
        resources:
          limits:
            cpu: 50m
            memory: 100Mi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 28086
          initialDelaySeconds: 15
          periodSeconds: 20
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
```