# acmesh

container image for acme.sh

## Packages

Check [GitHub Packages](https://github.com/guoyk93/acmesh/pkgs/container/acmesh)

## Usage

### Image

```
guoyk/acmesh:2023.6.12
ghcr.io/guoyk93/acmesh:2023.6.12
```

### Initialization

if `/data/.initialized` does not exist, this image will install `acme.sh` to `/data` directory, using environment variable `ACMESH_EMAIL` as email address.

It's strongly suggested to mount a `PersistentVolumeClaim` at `/data` to persist all you certs and configurations.

### CLI

Just execute `acme.sh` as usual.

### CronJob

`acme.sh` cronjob will execute at `15 1 * * *` automatically.

### Helper `acmesh-upload-qcloud`

You can use command `acmesh-upload-qcloud` to upload your certificate to Qcloud

**Usage**

```shell
export QCLOUD_SECRET_ID=xxxxxxxxxxxx
export QCLOUD_SECRET_KEY=xxxxxxxxxxxx
acmesh-upload-qcloud -domain mydomain.com
```

### Helper `acmesh-apply-secret`

You can use command `acmesh-apply-secret` to upload your certificate to Kubernetes cluster.

**Usage**

```shell
acmesh-apply-secret -domain mydomain.com -namespace my-namespace -name my-secret-name
```

Since this image is based on `minit`, you can write a unit file to patch secret periodically.

Example:

```yaml
# /etc/minit.d/upload-cert.yaml

kind: cron
name: acmesh-apply-external-cluster
cron: '25 1 * * *'
env:
  KUBECONFIG: /data/kubeconfigs/external.yaml
command:
  - acmesh-apply-secret
  - -domain
  - my-domain.com
  - -namespace
  - my-namespace
  - -name
  - my-secret-name
---
kind: cron
name: acmesh-apply-local-cluster
cron: '25 1 * * *'
command:
  - acmesh-patch-secret
  - -domain
  - my-domain.com
  - -namespace
  - my-namespace
  - -name
  - my-secret-name
```

View <https://github.com/guoyk93/minit> for detailed usage of `minit`

**RBAC Setup**

Here is a example to setup RBAC for `acmesh-apply-secret`.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: acmesh
automountServiceAccountToken: true
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: acmesh
rules:
  - verbs:
      - list
    apiGroups:
      - ''
    resources:
      - namespaces
  - verbs:
      - create
      - get
      - update
      - patch
    apiGroups:
      - ''
    resources:
      - secrets
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: acmesh
subjects:
  - kind: ServiceAccount
    name: acmesh
    namespace: 'YOUR NAMESPACE'
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: acmesh
```

## Donation

View <https://guoyk.net/donation>

## Credits

Guo Y.K., MIT License
