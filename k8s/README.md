# Local Kubernetes

This directory contains local-first Kubernetes manifests for learning `Deployment`, `Service`, `Ingress`, `ConfigMap`, `Secret`, and `PersistentVolumeClaim`.

The app images are local images:

- `spare-parts-api:local`
- `offer-price-worker:local`
- `spare-parts-provider:local`

Build and deploy:

```sh
make k8s-local-up
make k8s-status
```

`k8s-local-up` builds the local images, imports them into the local Kubernetes node containerd store, recreates the namespace, and waits for the deployments.

If your Kubernetes node container is not named `desktop-control-plane`, pass it explicitly:

```sh
make k8s-local-up K8S_NODE=<node-container-name>
```

Test with port-forward:

```sh
kubectl port-forward -n spare-parts svc/spare-parts-api 18080:8080
curl http://localhost:18080/spare-part/BRK-PAD-4521
```

Or test from inside the cluster:

```sh
kubectl run -n spare-parts curl-check --rm -i --restart=Never --image=busybox:1.36 -- wget -qO- http://spare-parts-api:8080/spare-part/BRK-PAD-4521
```

If you have an nginx ingress controller locally:

```sh
curl http://spare-parts.localhost/spare-part/BRK-PAD-4521
```

For a later AWS deployment, Postgres and Kafka should normally move out of these manifests and become managed services or Helm dependencies. The application manifests can then become Helm templates and Terraform can provision the cluster, ingress controller, managed database, and Kafka service.
