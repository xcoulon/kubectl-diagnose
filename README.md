# kubectl-diagnose (a.k.a. "Route 503")

[![GoDoc](https://godoc.org/github.com/xcoulon/kubectl-diagnose?status.svg)](https://godoc.org/github.com/xcoulon/kubectl-diagnose)
[![Go Report Card](https://goreportcard.com/badge/github.com/xcoulon/kubectl-diagnose)](https://goreportcard.com/report/github.com/xcoulon/kubectl-diagnose)
[![GitHub Action Build Status](https://github.com/xcoulon/kubectl-diagnose/workflows/ci-build/badge.svg)](https://github.com/xcoulon/kubectl-diagnose/actions?query=workflow%3Aci-build)
[![Codecov](https://codecov.io/gh/xcoulon/kubectl-diagnose/branch/main/graph/badge.svg)](https://codecov.io/gh/xcoulon/kubectl-diagnose)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

You just exposed your application running on OpenShift or Kubernetes to the outter world, but you're getting a `503 Service Unavailable` response when trying to access it?
Damn'it! 😬

Say hello to `kubectl-diagnose`👋

This tool attempts to find which of the `route`, `ingress`, `service`, `deployment`, `replicaset`, `statefulset`, `persistentvolumeclaim` or `pod` is misconfigured or failing, so you don't have to pull your hair out on such a problem.

## Installing

You can either:

- run `go install github.com/xcoulon/kubectl-diagnose` and add `$GOPATH/bin` to `$PATH` or move `$GOPATH/bin/kubectl-diagnose` in one of the directories of `$PATH`
- clone this repository, run `make install` and add `$GOPATH/bin` to `$PATH` or move `$GOPATH/bin/kubectl-terminate` in one of the directories of `$PATH`

## Supported Cases

`kubectl-diagnose` may detect:

- when the target service of an OpenShift route does not exist or has no matching port
- when there is no pod matching the service selector
- when there is no port in the containers matching the service target port
- when the target container is not running
- when the statefulset replicas or deployment replicas is set to `0`
- when a persistent volume claim is stuck in `pending` phase

## Usage

`kubectl-diagnose` can be run as a standalone command or as a plugin for `kubectl` and `oc` (for OpenShift 4 users) when the binary is in your `$PATH`.

`kubectl-diagnose` uses the same connection settings as the regular `kubectl` and `oc` commands:

- via the `--kubeconfig` flag if specified
- via `KUBECONFIG` env var if it exists
- from the default location (`$HOME/.kube/config` on Linux and macOS or `%USERPROFILE%\.kube\config` on Windows).

## Demo

The [examples directory](https://github.com/xcoulon/kubectl-diagnose/tree/main/examples) contains manifests with Routes, Services, Deployments and ConfigMaps for all supported cases, so you can try them by yourself on your OpenShift Cluster.

If you don't have an OpenShift cluster, you can sign into [Developer Sandbox](https://developers.redhat.com/developer-sandbox) and have a free access to an OpenShift cluster 30 days 😉

Using the `deployment-service-account-not-found.yaml` example:

```bash
$ oc apply -f examples/deployment-service-account-not-found.yaml
route.route.openshift.io/deploy-sa-notfound created
service/deploy-sa-notfound created
deployment.apps/deploy-sa-notfound created
configmap/caddy-config-port-8080 configured

$ oc diagnose route/deploy-sa-notfound
👀 checking route 'deploy-sa-notfound' in namespace 'test'...
👀 checking service 'deploy-sa-notfound' in namespace 'test'...
👀 checking replicaset 'deploy-sa-notfound-59b5d8468f' in namespace 'test'...
👻 replicaset 'deploy-sa-notfound-59b5d8468f' failed to create pods: pods "deploy-sa-notfound-59b5d8468f-" is forbidden: error looking up service account test/deploy-sa-notfound: serviceaccount "deploy-sa-notfound" not found
```

## Contributing

Feel free to [open an issue](https://github.com/xcoulon/kubectl-diagnose/issues) describing a case that is not covered yet. It could help you and others in the future!🤩
If you're in the mood to implement it, that's awesome! 🙌
You can submit a pull-request along with tests written with Ginkgo and a ideally, title based on the [Conventional Commits specification](https://www.conventionalcommits.org/en/v1.0.0/).

Don't get discouraged if you don't get an immediate response, this is a side-project 😅

## License

`kubectl-diagnose` is available under the terms of the [Apache License 2.0](https://raw.githubusercontent.com/xcoulon/kubectl-diagnose/LICENSE).
