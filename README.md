# acmesh

container image for acme.sh

## Packages

Check [GitHub Packages](https://github.com/guoyk93/acmesh/pkgs/container/acmesh)

## Usage

### Initialization

if `/data/.initialized` does not exist, this image will install `acme.sh` to `/data` directory, using environment variable `ACMESH_EMAIL` as email address.

It's strongly suggested to mount a `PersistentVolumeClaim` at `/data` to persist all you certs and configurations.

### CLI

Just execute `acme.sh` as usual.

### CronJob

`acme.sh` cronjob will execute at `15 1 * * *` automatically.

### Helper Scripts

You can use script `acmesh-patch-secret` to upload your certificate to Kubernetes cluster.

```shell
acmesh-patch-secret mydomain.com my-namespace my-secret-name
```

Since this image is based on `minit`, you can write a unit file to patch secret periodically.

Example:

```yaml
# /etc/minit.d/upload-cert.yaml

kind: cron
name: acmesh-patch-external-cluster
cron: '25 1 * * *'
env:
  KUBECONFIG: /data/kubeconfigs/external.yaml
command:
  - acmesh-patch-secret
  - my-domain.com
  - my-namespace
  - my-secret-name
---
kind: cron
name: acmesh-patch-local-cluster
cron: '25 1 * * *'
command:
  - acmesh-patch-secret
  - my-domain.com
  - my-namespace
  - my-secret-name
```

View <https://github.com/guoyk93/minit> for detailed usage of `minit`

## Donation

View <https://guoyk.net/donation>

## Credits

Guo Y.K., MIT License
