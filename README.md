# gardener-extension-accounting

Deploys cluster accounting components into the seed's shoot namespaces.

## Deploying into local Gardener

It is possible to deploy gardener-extension-accounting to a local Gardener cluster.
Currently only a [patched version v1.56.2 of Gardener](https://github.com/metal-stack/gardener/tree/release-v1.56.2%2BupdatedCerts) is supported.
Clone it and spin up the local development environment:

```bash
git clone git@github.com:metal-stack/gardener.git --branch release-v1.56.2+updatedCerts
cd gardener
make kind-up gardener-up
```

Now point your `KUBECONFIG` to the Gardener cluster:

```bash
export KUBECONFIG=/path/to/gardener/example/gardener-local/kind/kubeconfig
```

Next, we need to deploy some CRDs for cluster-wide network policies of the firewall controller:

```bash
git clone git@github.com:metal-stack/firewall-controller.git
kubectl apply -f config/crd/bases/metal-stack.io_clusterwidenetworkpolicies.yaml
```

Now we create the `firewall` namespace, as the accounting extension tries to deploy a cluster-wide network policy in there:

```bash
kubectl create ns firewall
```

Now you are able to deploy the accounting extension itself:

```bash
kubectl apply -k example/
kubectl apply -f example/shoot.yaml
make push-to-gardener-local
```

Finally you need to connect your local gardener installation with a running metal-stack. Though this couldn't be done locally due to performance reasons.
