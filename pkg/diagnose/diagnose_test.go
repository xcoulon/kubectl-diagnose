package diagnose_test

import (
	"fmt"
	"io"
	"time"

	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	. "github.com/xcoulon/kubectl-diagnose/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("container in CrashLoopBackOff status",
	func(kind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/pod-crash-loop-back-off.yaml", "resources/pod-crash-loop-back-off.logs")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		if diagnose.IsRoute(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'crash-loop-back-off' in namespace 'test'...`))
		}
		if diagnose.IsRoute(kind) || diagnose.IsService(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'crash-loop-back-off' in namespace 'test'...`))
		}
		if diagnose.IsDeployment(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking deployment 'crash-loop-back-off' in namespace 'test'...`))
		}
		if diagnose.IsDeployment(kind) || diagnose.IsReplicaSet(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'crash-loop-back-off-7994787459' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'crash-loop-back-off-7994787459-2nrz5' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		// event
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod events...`))
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-12T18:02:28Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Back-off restarting failed container`, time.Since(lastTimestamp).Truncate(time.Second))))
		// logs
		Expect(logger.Output()).To(ContainSubstring(`👀 checking 'default' container logs...`))
		Expect(logger.Output()).To(ContainSubstring(`🗒  Error: loading initial config: loading new config: http app module: start: listening on :80: listen tcp :80: bind: permission denied`))
	},
	Entry("should detect from pod", "pod", "test", "crash-loop-back-off-7994787459-2nrz5"),
	Entry("should detect from replicaset", "replicaset", "test", "crash-loop-back-off-7994787459"),
	Entry("should detect from deployment", "deployment", "test", "crash-loop-back-off"),
	Entry("should detect from service", "service", "test", "crash-loop-back-off"),
	Entry("should detect from route", "route", "test", "crash-loop-back-off"),
)

var _ = DescribeTable("container in ImagePullBackOff status",
	func(kind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/pod-image-pull-back-off.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		if diagnose.IsRoute(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'image-pull-back-off' in namespace 'test'...`))
		}
		if diagnose.IsRoute(kind) || diagnose.IsService(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'image-pull-back-off' in namespace 'test'...`))
		}
		if diagnose.IsDeployment(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking deployment 'image-pull-back-off' in namespace 'test'...`))
		}
		if diagnose.IsDeployment(kind) || diagnose.IsReplicaSet(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'image-pull-back-off-9bbb4f9bd' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'ImagePullBackOff': Back-off pulling image "unknown:v0.0.0"`))
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-13T07:59:04Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Error: ImagePullBackOff`, time.Since(lastTimestamp).Truncate(time.Second))))
	},
	Entry("should detect from pod", "pod", "test", "image-pull-back-off-9bbb4f9bd-pjj55"),
	Entry("should detect from replicaset", "replicaset", "test", "image-pull-back-off-9bbb4f9bd"),
	Entry("should detect from deployment", "deployment", "test", "image-pull-back-off"),
	Entry("should detect from service", "service", "test", "image-pull-back-off"),
	Entry("should detect from route", "route", "test", "image-pull-back-off"),
)

var _ = DescribeTable("container with unknown configmap mount",
	func(kind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/pod-unknown-configmap.yaml") // no logs, container is not created
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		if diagnose.IsRoute(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'unknown-configmap' in namespace 'test'...`))
		}
		if diagnose.IsRoute(kind) || diagnose.IsService(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'unknown-configmap' in namespace 'test'...`))
		}
		if diagnose.IsDeployment(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking deployment 'unknown-configmap' in namespace 'test'...`))
		}
		if diagnose.IsDeployment(kind) || diagnose.IsReplicaSet(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'unknown-configmap-76476b7d5' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'unknown-configmap-76476b7d5-q2khp' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'ContainerCreating'`))
		Expect(logger.Output()).NotTo(ContainSubstring(`👻 container 'default' is waiting with reason 'ContainerCreating':`)) // ensure there is no `:` followed by an empty message
		// no logs: container has not started
		// events
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod events...`))
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-13T17:19:34Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Unable to attach or mount volumes: unmounted volumes=[caddy-config], unattached volumes=[caddy-config caddy-config-cache kube-api-access-62xrc]: timed out waiting for the condition`, time.Since(lastTimestamp).Truncate(time.Second))))
	},
	Entry("should detect from pod", "pod", "test", "unknown-configmap-76476b7d5-q2khp"),
	Entry("should detect from replicaset", "replicaset", "test", "unknown-configmap-76476b7d5"),
	Entry("should detect from deployment", "deployment", "test", "unknown-configmap"),
	Entry("should detect from service", "service", "test", "unknown-configmap"),
	Entry("should detect from route", "route", "test", "unknown-configmap"),
)

var _ = DescribeTable("container with readiness probe error",
	func(kind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/pod-readiness-probe-error.yaml", "resources/pod-readiness-probe-error.logs")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		if diagnose.IsRoute(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'readiness-probe-error' in namespace 'test'...`))
		}
		if diagnose.IsRoute(kind) || diagnose.IsService(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'readiness-probe-error' in namespace 'test'...`))
		}
		if diagnose.IsDeployment(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking deployment 'readiness-probe-error' in namespace 'test'...`))
		}
		if diagnose.IsDeployment(kind) || diagnose.IsReplicaSet(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'readiness-probe-error-6cb7664768' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'readiness-probe-error-6cb7664768-qlmns' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		// events
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-13T21:55:27Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Readiness probe failed: HTTP probe failed with statuscode: 404`, time.Since(lastTimestamp).Truncate(time.Second))))
		// logs
		Expect(logger.Output()).To(ContainSubstring(`👀 checking 'default' container logs...`))
		Expect(logger.Output()).To(ContainSubstring("🤷 no 'error'/'fatal'/'panic'/'emerg' messages found in the container logs"))
	},
	Entry("should detect from pod", "pod", "test", "readiness-probe-error-6cb7664768-qlmns"),
	Entry("should detect from replicaset", "replicaset", "test", "readiness-probe-error-6cb7664768"),
	Entry("should detect from deployment", "deployment", "test", "readiness-probe-error"),
	Entry("should detect from service", "service", "test", "readiness-probe-error"),
	Entry("should detect from route", "route", "test", "readiness-probe-error"),
)

var _ = DescribeTable("serviceaccount not found",
	func(kind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/deployment-service-account-not-found.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		if diagnose.IsRoute(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'sa-notfound' in namespace 'test'...`))
		}
		if diagnose.IsRoute(kind) || diagnose.IsService(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'sa-notfound' in namespace 'test'...`))
		}
		if diagnose.IsDeployment(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking deployment 'sa-notfound' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'sa-notfound-59b5d8468f' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 replicaset 'sa-notfound-59b5d8468f' failed to create pods: pods "sa-notfound-59b5d8468f-" is forbidden: error looking up service account test/sa-notfound: serviceaccount "sa-notfound" not found`))
	},
	Entry("should detect from replicaset", "replicaset", "test", "sa-notfound-59b5d8468f"),
	Entry("should detect from deployment", "deployment", "test", "sa-notfound"),
	Entry("should detect from service", "service", "test", "sa-notfound"),
	Entry("should detect from route", "route", "test", "sa-notfound"),
)

var _ = DescribeTable("should detect no matching pods",
	func(kind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/service-no-matching-pods.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		if diagnose.IsRoute(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'no-matching-pods' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'no-matching-pods' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no pods matching label selector 'app=invalid' found in namespace 'test'`))
		Expect(logger.Output()).To(ContainSubstring(`💡 you may want to verify that the pods exist and their labels match 'app=invalid'`))
	},
	Entry("should detect from route", "route", "test", "no-matching-pods"),
	Entry("should detect from service", "service", "test", "no-matching-pods"),
)

var _ = DescribeTable("should detect invalid service target port as string",
	func(kind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/service-invalid-target-port-str.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		if diagnose.IsRoute(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'invalid-service-target-port-str' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'invalid-service-target-port-str' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no container with port 'https' in pod 'invalid-service-target-port-str-76d5db5c9b-s8wpq'`))
	},
	Entry("should detect from route", "route", "test", "invalid-service-target-port-str"),
	Entry("should detect from service", "service", "test", "invalid-service-target-port-str"),
)

var _ = DescribeTable("should detect invalid service target port as int",
	func(kind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/service-invalid-target-port-int.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		if diagnose.IsRoute(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'invalid-service-target-port-int' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'invalid-service-target-port-int' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no container with port '8443' in pod 'invalid-service-target-port-int-bbcb4fd5d-k8kg8'`))
	},
	Entry("should detect from route", "route", "test", "invalid-service-target-port-int"),
	Entry("should detect from service", "service", "test", "invalid-service-target-port-int"),
)

var _ = DescribeTable("should detect missing route target service",
	func(kind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/route-unknown-target-service.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'unknown-target-service' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring("👻 unable to find service 'unknown'"))
	},
	Entry("should detect from route", "route", "test", "unknown-target-service"),
)

var _ = DescribeTable("should detect invalid route target port as string",
	func(kind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/route-invalid-target-port-str.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'invalid-route-target-port-str' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring("👻 route target port 'https' is not defined in service 'invalid-route-target-port-str'"))
	},
	Entry("should detect from route", "route", "test", "invalid-route-target-port-str"),
)

var _ = DescribeTable("should detect invalid route target port as int",
	func(kind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/route-invalid-target-port-int.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'invalid-route-target-port-int' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring("👻 route target port '8443' is not defined in service 'invalid-route-target-port-int'"))
	},
	Entry("should detect from route", "route", "test", "invalid-route-target-port-int"),
)

var _ = DescribeTable("should detect zero replicas specified in deployment",
	func(kind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/deployment-zero-replica.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		if diagnose.IsRoute(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'zero-replica' in namespace 'test'...`))
		}
		if diagnose.IsRoute(kind) || diagnose.IsService(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'zero-replica' in namespace 'test'...`))
		}
		if diagnose.IsRoute(kind) || diagnose.IsService(kind) || diagnose.IsReplicaSet(kind) {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'zero-replica-9bccf7d88' in namespace 'test'...`))
		}
		// in all cases
		Expect(logger.Output()).To(ContainSubstring(`👀 checking deployment 'zero-replica' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 number of desired replicas for deployment 'zero-replica' is set to 0`))
		Expect(logger.Output()).To(ContainSubstring(`💡 you may want run 'oc scale --replicas=1 deployment/zero-replica -n test' or increase the 'replicas' value in the deployment specs`))
	},
	Entry("should detect from route", "route", "test", "zero-replica"),
	Entry("should detect from service", "service", "test", "zero-replica"),
	Entry("should detect from deployment", "deployment", "test", "zero-replica"),
	Entry("should detect from replicaset", "replicaset", "test", "zero-replica-9bccf7d88"),
)
