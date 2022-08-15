## test

```bash
while true; do curl -s https://service-leader-election.dev.alldigitalads.com/debug && echo; sleep 1; done
```

## rbac

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
- apiGroups: ["coordination.k8s.io"]
  resources: ["leases"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: service-leader-election-rolebinding
roleRef:
  kind: Role
  name: service-leader-election-role
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  name: service-leader-election
  namespace: {{ .Release.Namespace }}
```