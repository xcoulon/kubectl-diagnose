package diagnose_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"
	"github.com/xcoulon/kubectl-diagnose/testsupport"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestDiagnose(t *testing.T) {

	// given
	now := time.Now()

	type entrypoint struct {
		kind      diagnose.ResourceKind
		namespace string
		name      string
	}
	testdata := map[string][]struct {
		title         string
		resources     []string
		entrypoints   []entrypoint
		expectedFound bool
		expectedMsgs  []string
	}{
		// --------------------------------------------------------
		// Diagnose errors on Routes
		// --------------------------------------------------------
		"routes": {
			{
				title: "should detect missing route target service",
				resources: []string{
					"resources/route-unknown-target-service.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "unknown-target-service",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					"👻 unable to find service 'unknown'",
				},
			},
			{
				title: "should detect invalid route target port as string",
				resources: []string{
					"resources/route-invalid-target-port-str.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "invalid-route-target-port-str",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					"👻 route target port 'https' is not defined in service 'invalid-route-target-port-str'",
				},
			},
			{
				title: "should detect invalid route target port as int",
				resources: []string{
					"resources/route-invalid-target-port-int.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "invalid-route-target-port-int",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					"👻 route target port '8443' is not defined in service 'invalid-route-target-port-int'",
				},
			},
		},

		// --------------------------------------------------------
		// Diagnose errors on Ingresses
		// --------------------------------------------------------
		"ingresses": {
			{
				title: "should detect missing target service",
				resources: []string{
					"resources/ingress-unknown-target-service.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Ingress,
						namespace: "test",
						name:      "unknown-target-service",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					"👻 unable to find service 'unknown' associated with host 'unknown-target-service.test' and path '/'",
				},
			},

			{
				title: "should detect invalid service port",
				resources: []string{
					"resources/ingress-invalid-service-port.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Ingress,
						namespace: "test",
						name:      "invalid-service-port",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					"👻 port '8081' is not defined in service 'invalid-service-port'",
				},
			},
			{
				title: "should detect invalid service name",
				resources: []string{
					"resources/ingress-invalid-service-name.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Ingress,
						namespace: "test",
						name:      "invalid-service-name",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					"👻 port 'https' is not defined in service 'invalid-service-name'",
				},
			},
			{
				title: "should detect invalid ingressclassname",
				resources: []string{
					"resources/ingress-invalid-ingressclassname.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Ingress,
						namespace: "test",
						name:      "invalid-ingressclassname",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					"👻 unable to find ingressclass 'invalid'",
				},
			},
			{
				title: "should not fail when get ingressclass is forbidden",
				resources: []string{
					"resources/ingress-forbidden-ingressclassname.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Ingress,
						namespace: "test",
						name:      "forbidden-ingressclassname",
					},
				},
				expectedFound: false,
				expectedMsgs: []string{
					"🤷 unable to verify ingressclass 'forbidden': ingressclass 'forbidden' is forbidden: User cannot get ingressclass resources at the cluster level (get ingressclasses.networking.k8s.io forbidden",
				},
			},
		},

		// --------------------------------------------------------
		// Diagnose errors on Services
		// --------------------------------------------------------
		"services": {
			{
				title: "should detect no matching pods",
				resources: []string{
					"resources/service-no-matching-pods.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "no-matching-pods",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "no-matching-pods",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`👻 no pods matching label selector 'app=invalid' found in namespace 'test'`,
					`💡 you may want to verify that the pods exist and their labels match 'app=invalid'`,
				},
			},
			{
				title: "should detect invalid service target port as string",
				resources: []string{
					"resources/service-invalid-target-port-str.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "invalid-service-target-port-str",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "invalid-service-target-port-str",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`👻 no container with port 'https' in pod 'invalid-service-target-port-str-76d5db5c9b-s8wpq'`,
				},
			},
			{
				title: "should detect invalid service target port as int",
				resources: []string{
					"resources/service-invalid-target-port-int.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "invalid-service-target-port-int",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "invalid-service-target-port-int",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`👻 no container with port '8443' in pod 'invalid-service-target-port-int-bbcb4fd5d-k8kg8'`,
				},
			},
		},
		// --------------------------------------------------------
		// Diagnose errors on Deployments / ReplicaSets
		// --------------------------------------------------------,
		"deployments": {
			{
				title: "should detect zero replicas specified in deployment",
				resources: []string{
					"resources/deployment-zero-replica.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.ReplicaSet,
						namespace: "test",
						name:      "deploy-zero-replica-9bccf7d88",
					},
					{
						kind:      diagnose.Deployment,
						namespace: "test",
						name:      "deploy-zero-replica",
					},
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "deploy-zero-replica",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "deploy-zero-replica",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`👻 number of desired replicas for deployment 'deploy-zero-replica' is set to 0`,
					`💡 run 'oc scale --replicas=1 deployment/deploy-zero-replica -n test' or increase the 'replicas' value in the deployment specs`,
				},
			},
			{
				title: "should detect invalid serviceaccount specified in deployment",
				resources: []string{
					"resources/deployment-service-account-not-found.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.ReplicaSet,
						namespace: "test",
						name:      "deploy-sa-notfound-59b5d8468f",
					},
					{
						kind:      diagnose.Deployment,
						namespace: "test",
						name:      "deploy-sa-notfound",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "deploy-sa-notfound",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "deploy-sa-notfound",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`👻 replicaset 'deploy-sa-notfound-59b5d8468f' failed to create pods: pods "deploy-sa-notfound-59b5d8468f-" is forbidden: error looking up service account test/deploy-sa-notfound: serviceaccount "deploy-sa-notfound" not found`,
				},
			},
			{
				title: "should detect invalid serviceaccount specified in deployment with multiple replicasets",
				resources: []string{
					"resources/deployment-multiple-replicasets-failedcreate.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.ReplicaSet,
						namespace: "test",
						name:      "deploy-multiple-rs-c5d7d87f",
					},
					{
						kind:      diagnose.Deployment,
						namespace: "test",
						name:      "deploy-multiple-rs",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					"👻 containers with unready status: [kube-rbac-proxy default]",
					"👻 container 'default' is waiting with reason 'ContainerCreating'",
					"👻 container 'kube-rbac-proxy' is waiting with reason 'ContainerCreating'",
				},
			},
		},

		// --------------------------------------------------------
		// Diagnose errors StatefulSets
		// --------------------------------------------------------
		"statefulsets": {
			{
				title: "should detect zero replicas specified in statefulset",
				resources: []string{
					"resources/statefulset-zero-replica.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.StatefulSet,
						namespace: "test",
						name:      "sts-zero-replica",
					},
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "sts-zero-replica",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "sts-zero-replica",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`👻 number of desired replicas for statefulset 'sts-zero-replica' is set to 0`,
					`💡 run 'oc scale --replicas=1 sts/sts-zero-replica -n test' or increase the 'replicas' value in the statefulset specs`,
				},
			},
			{
				title: "should detect invalid serviceaccount specified in statefulset",
				resources: []string{
					"resources/statefulset-service-account-not-found.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.StatefulSet,
						namespace: "test",
						name:      "sts-sa-notfound",
					},
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "sts-sa-notfound",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "sts-sa-notfound",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					fmt.Sprintf(`⚡️ %s ago: FailedCreate: create Pod sts-sa-notfound-0 in StatefulSet sts-sa-notfound failed error: pods "sts-sa-notfound-0" is forbidden: error looking up service account test/unknown: serviceaccount "unknown" not found`, since(now, "2022-11-27T08:51:34Z")),
				},
			},
			{
				title: "should detect invalid storageclass specified in statefulset",
				resources: []string{
					"resources/statefulset-invalid-storageclass.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Pod,
						namespace: "test",
						name:      "sts-invalid-sc-0",
					},
					{
						kind:      diagnose.StatefulSet,
						namespace: "test",
						name:      "sts-invalid-sc",
					},
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "sts-invalid-sc",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "sts-invalid-sc",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					fmt.Sprintf(`⚡️ %s ago: FailedScheduling: 0/12 nodes are available: 12 pod has unbound immediate PersistentVolumeClaims. preemption: 0/12 nodes are available: 12 Preemption is not helpful for scheduling.`, since(now, "2022-11-26T08:40:16.475828Z")),
					fmt.Sprintf(`⚡️ %s ago: ProvisioningFailed: storageclass.storage.k8s.io "unknown" not found`, since(now, "2022-11-26T09:40:20Z")),
				},
			},
		},

		// --------------------------------------------------------
		// Diagnose errors Pods
		// --------------------------------------------------------
		"pods": {
			{
				title: "should detect default container in CrashLoopBackOff status from pod",
				resources: []string{
					"resources/deployment-pod-crash-loop-back-off.yaml",
					"resources/deployment-pod-crash-loop-back-off.logs",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Pod,
						namespace: "test",
						name:      "deploy-crash-loop-back-off-7994787459-2nrz5",
					},
					{
						kind:      diagnose.ReplicaSet,
						namespace: "test",
						name:      "deploy-crash-loop-back-off-7994787459",
					},
					{
						kind:      diagnose.Deployment,
						namespace: "test",
						name:      "deploy-crash-loop-back-off",
					},
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "deploy-crash-loop-back-off",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "deploy-crash-loop-back-off",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`👻 containers with unready status: [default]`,
					fmt.Sprintf(`⚡️ %s ago: BackOff: Back-off restarting failed container`, since(now, "2022-11-12T18:02:28Z")),
					`🗒  Error: loading initial config: loading new config: http app module: start: listening on :80: listen tcp :80: bind: permission denied`,
				},
			},
			{
				title: "should detect proxy container in CrashLoopBackOff status",
				resources: []string{
					"resources/deployment-pod-crash-loop-back-off-proxy.yaml",
					"resources/deployment-pod-crash-loop-back-off-proxy.logs",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Pod,
						namespace: "test",
						name:      "caddy-76c8d8fdfb-qgssh",
					},
					{
						kind:      diagnose.ReplicaSet,
						namespace: "test",
						name:      "caddy-76c8d8fdfb",
					},
					{
						kind:      diagnose.Deployment,
						namespace: "test",
						name:      "caddy",
					},
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "caddy",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`👻 containers with unready status: [kube-rbac-proxy]`,
					`🗒  FLAG: --oidc-username-claim="email"`,
					`  E0106 06:27:45.761479       1 run.go:74] "command failed" err="failed to read the config file: failed to read resource-attribute file: open /etc/kube-rbac-proxy/config.yaml: no such file or directory"`,
					fmt.Sprintf(`⚡️ %s ago: BackOff: Back-off restarting failed container`, since(now, "2023-01-04T06:59:16Z")),
				},
			},
			{
				title: "should detect container in ImagePullBackOff status",
				resources: []string{
					"resources/deployment-pod-image-pull-back-off.yaml",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Pod,
						namespace: "test",
						name:      "deploy-image-pull-back-off-9bbb4f9bd-pjj55",
					},
					{
						kind:      diagnose.ReplicaSet,
						namespace: "test",
						name:      "deploy-image-pull-back-off-9bbb4f9bd",
					},
					{
						kind:      diagnose.Deployment,
						namespace: "test",
						name:      "deploy-image-pull-back-off",
					},
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "deploy-image-pull-back-off",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`👻 containers with unready status: [default]`,
					`👻 container 'default' is waiting with reason 'ImagePullBackOff': Back-off pulling image "unknown:v0.0.0"`,
				},
			},
			{
				title: "should detect container with readiness probe error",
				resources: []string{
					"resources/deployment-pod-readiness-probe-error.yaml",
					"resources/deployment-pod-readiness-probe-error.logs",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Pod,
						namespace: "test",
						name:      "deploy-readiness-probe-error-6cb7664768-qlmns",
					},
					{
						kind:      diagnose.ReplicaSet,
						namespace: "test",
						name:      "deploy-readiness-probe-error-6cb7664768",
					},
					{
						kind:      diagnose.Deployment,
						namespace: "test",
						name:      "deploy-readiness-probe-error",
					},
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "deploy-readiness-probe-error",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "deploy-readiness-probe-error",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`👻 containers with unready status: [default]`,
					fmt.Sprintf(`⚡️ %s ago: Unhealthy: Readiness probe failed: HTTP probe failed with statuscode: 404`, since(now, "2022-11-13T21:55:27Z")),
					"🤷 no 'error'/'failed'/'fatal'/'panic'/'emerg' messages found in the 'default' container logs",
				},
			},
			{
				title: "should detect container with unknown configmap mount",
				resources: []string{
					"resources/deployment-pod-unknown-configmap.yaml", // no logs, container is not created
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Pod,
						namespace: "test",
						name:      "deploy-unknown-cm-76476b7d5-q2khp",
					},
					{
						kind:      diagnose.ReplicaSet,
						namespace: "test",
						name:      "deploy-unknown-cm-76476b7d5",
					},
					{
						kind:      diagnose.Deployment,
						namespace: "test",
						name:      "deploy-unknown-cm",
					},
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "deploy-unknown-cm",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "deploy-unknown-cm",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`👻 containers with unready status: [default]`,
					`👻 container 'default' is waiting with reason 'ContainerCreating'`,
					fmt.Sprintf(`⚡️ %s ago: FailedMount: Unable to attach or mount volumes: unmounted volumes=[caddy-config], unattached volumes=[caddy-config caddy-config-cache kube-api-access-62xrc]: timed out waiting for the condition`, since(now, "2022-11-13T17:19:34Z")),
				},
			},
			{
				title: "should detect container with unknown configmap mount",
				resources: []string{
					"resources/statefulset-pod-unknown-configmap.yaml", // no logs, container is not created
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.StatefulSet,
						namespace: "test",
						name:      "sts-unknown-cm",
					},
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "sts-unknown-cm",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "sts-unknown-cm",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`👻 containers with unready status: [default]`,
					`👻 container 'default' is waiting with reason 'CreateContainerConfigError': configmap "sts-unknown-cm" not found`,
					fmt.Sprintf(`⚡️ %s ago: Failed: Error: configmap "sts-unknown-cm" not found`, since(now, "2022-12-01T05:40:55Z")),
				},
			},
			{
				title: "should detect container not ready but starting",
				resources: []string{
					"resources/pod-container-starting-not-ready.yaml", // container is taking time to start
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Pod,
						namespace: "test",
						name:      "prometheus-container-starting",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`👻 containers with unready status: [prometheus]`,
				},
			},
			{
				title: "should detect error logs in target container",
				resources: []string{
					"resources/pod-container-with-errors.yaml",
					"resources/pod-container-with-errors.logs",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "alertmanager-oauth2",
					},
				},
				expectedFound: true,
				expectedMsgs: []string{
					`tls: bad record MAC`,
					`wrong TLS secret mounted on the 'oauth2-proxy' container?`,
				},
			},
		},
		// --------------------------------------------------------
		// Diagnose no errors when all good
		// --------------------------------------------------------

		"all good": {
			{
				title: "should not find errors",
				resources: []string{
					"resources/all-good.yaml",
					"resources/all-good.logs",
				},
				entrypoints: []entrypoint{
					{
						kind:      diagnose.Pod,
						namespace: "test",
						name:      "all-good-785d8bcc5f-x85p2",
					},
					{
						kind:      diagnose.ReplicaSet,
						namespace: "test",
						name:      "all-good-785d8bcc5f",
					},
					{
						kind:      diagnose.Service,
						namespace: "test",
						name:      "all-good",
					},
					{
						kind:      diagnose.Route,
						namespace: "test",
						name:      "all-good",
					},
				},
				expectedFound: false,
				expectedMsgs: []string{
					diagnose.NotFoundMsg,
				},
			},
		},
	}

	for category, tcs := range testdata {
		t.Run(category, func(t *testing.T) {
			for _, tc := range tcs {
				t.Run(tc.title, func(t *testing.T) {
					for _, e := range tc.entrypoints {
						t.Run(fmt.Sprintf("from %v", e.kind), func(t *testing.T) {
							// given
							buf := &strings.Builder{}
							logger := log.NewWithOptions(buf, log.Options{
								TimeFormat: time.Kitchen,
								Level:      log.DebugLevel,
								Formatter:  log.TextFormatter,
							})
							apiserver, err := testsupport.NewFakeAPIServer(logger, tc.resources...)
							require.NoError(t, err)
							cfg := testsupport.NewConfig(apiserver.URL, "/api")
							ctx := context.WithValue(context.TODO(), diagnose.NowContextKey, now)
							// when
							found, err := diagnose.Diagnose(ctx, logger, cfg, e.kind, e.namespace, e.name)

							// then
							require.NoError(t, err)
							assert.Equal(t, tc.expectedFound, found)
							for _, m := range tc.expectedMsgs {
								assert.Contains(t, buf.String(), m)
							}
						})
					}
				})
			}
		})
	}

	// --------------------------------------------------------
	// Server-side Errors
	// --------------------------------------------------------

	t.Run("should handle internal server errors", func(t *testing.T) {

		entrypoints := []entrypoint{
			{
				kind:      diagnose.Pod,
				namespace: "test",
				name:      "error",
			},
			{
				kind:      diagnose.PersistentVolumeClaim,
				namespace: "test",
				name:      "error",
			},
			{
				kind:      diagnose.StatefulSet,
				namespace: "test",
				name:      "error",
			},
			{
				kind:      diagnose.Deployment,
				namespace: "test",
				name:      "error",
			},
			{
				kind:      diagnose.Service,
				namespace: "test",
				name:      "error",
			},
			{
				kind:      diagnose.Route,
				namespace: "test",
				name:      "error",
			},
		}

		for _, e := range entrypoints {
			t.Run(fmt.Sprintf("from %v", e.kind), func(t *testing.T) {
				// given
				logger := log.NewWithOptions(os.Stdout, log.Options{
					TimeFormat: time.Kitchen,
					Level:      log.DebugLevel,
				})

				apiserver, err := testsupport.NewFakeAPIServer(logger)
				require.NoError(t, err)
				cfg := testsupport.NewConfig(apiserver.URL, "/api")

				// when
				_, err = diagnose.Diagnose(context.TODO(), logger, cfg, e.kind, e.namespace, e.name)

				// then
				assert.True(t, apierrors.IsInternalError(err))
			})
		}
	})

	t.Run("should handle not found errors", func(t *testing.T) {

		entrypoints := []entrypoint{
			{
				kind:      diagnose.Pod,
				namespace: "test",
				name:      "notfound",
			},
			{
				kind:      diagnose.PersistentVolumeClaim,
				namespace: "test",
				name:      "notfound",
			},
			{
				kind:      diagnose.StatefulSet,
				namespace: "test",
				name:      "notfound",
			},
			{
				kind:      diagnose.Deployment,
				namespace: "test",
				name:      "notfound",
			},
			{
				kind:      diagnose.Service,
				namespace: "test",
				name:      "notfound",
			},
			{
				kind:      diagnose.Route,
				namespace: "test",
				name:      "notfound",
			},
		}

		for _, e := range entrypoints {
			t.Run(fmt.Sprintf("from %v", e.kind), func(t *testing.T) {
				// given
				logger := log.NewWithOptions(os.Stdout, log.Options{
					TimeFormat: time.Kitchen,
					Level:      log.DebugLevel,
				})

				apiserver, err := testsupport.NewFakeAPIServer(logger)
				require.NoError(t, err)
				cfg := testsupport.NewConfig(apiserver.URL, "/api")

				// when
				_, err = diagnose.Diagnose(context.TODO(), logger, cfg, e.kind, e.namespace, e.name)

				// then
				assert.True(t, apierrors.IsNotFound(err))
			})
		}
	})
}

func since(now time.Time, tss string) time.Duration {
	ts, _ := time.Parse("2006-01-02T15:04:05Z", tss)
	return now.Sub(ts).Truncate(time.Second)
}
