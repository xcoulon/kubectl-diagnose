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
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		// event
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod events...`))
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-12T18:02:28Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Back-off restarting failed container`, time.Since(lastTimestamp).Truncate(time.Second))))
		// logs
		Expect(logger.Output()).To(ContainSubstring(`👀 checking logs in 'default' container...`))
		Expect(logger.Output()).To(ContainSubstring(`🗒  Error: loading initial config: loading new config: http app module: start: listening on :80: listen tcp :80: bind: permission denied`))
	},
	Entry("should detect from pod", diagnose.Pod, "test", "crash-loop-back-off-7994787459-2nrz5"),
	Entry("should detect from replicaset", diagnose.ReplicaSet, "test", "crash-loop-back-off-7994787459"),
	Entry("should detect from service", diagnose.Service, "test", "crash-loop-back-off"),
	Entry("should detect from route", diagnose.Route, "test", "crash-loop-back-off"),
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
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'ImagePullBackOff': Back-off pulling image "unknown:v0.0.0"`))
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-13T07:59:04Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Error: ImagePullBackOff`, time.Since(lastTimestamp).Truncate(time.Second))))
		Expect(logger.Output()).To(ContainSubstring(`👀 checking logs in 'default' container...`))
		Expect(logger.Output()).To(ContainSubstring(`🤷 no relevant message found in the pod logs (but you may want to check yourself)`))
	},
	Entry("should detect from pod", diagnose.Pod, "test", "image-pull-back-off-9bbb4f9bd-pjj55"),
	Entry("should detect from replicaset", diagnose.ReplicaSet, "test", "image-pull-back-off-9bbb4f9bd"),
	Entry("should detect from service", diagnose.Service, "test", "image-pull-back-off"),
	Entry("should detect from route", diagnose.Route, "test", "image-pull-back-off"),
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
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'unknown-configmap-76476b7d5-q2khp'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'ContainerCreating'`))
		Expect(logger.Output()).NotTo(ContainSubstring(`👻 container 'default' is waiting with reason 'ContainerCreating':`)) // ensure there is no `:` followed by an empty message
		// no logs: container has not started
		// events
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod events...`))
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-13T17:19:34Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Unable to attach or mount volumes: unmounted volumes=[caddy-config], unattached volumes=[caddy-config caddy-config-cache kube-api-access-62xrc]: timed out waiting for the condition`, time.Since(lastTimestamp).Truncate(time.Second))))
	},
	Entry("should detect from pod", diagnose.Pod, "test", "unknown-configmap-76476b7d5-q2khp"),
	Entry("should detect from replicaset", diagnose.ReplicaSet, "test", "unknown-configmap-76476b7d5"),
	Entry("should detect from service", diagnose.Service, "test", "unknown-configmap"),
	Entry("should detect from route", diagnose.Route, "test", "unknown-configmap"),
)

var _ = DescribeTable("container with readiness probe error",
	func(kind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/pod-readiness-probe-error.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'readiness-probe-error-6cb7664768-qlmns'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		// events
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-13T21:55:27Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Readiness probe failed: HTTP probe failed with statuscode: 404`, time.Since(lastTimestamp).Truncate(time.Second))))
	},
	Entry("should detect from pod", diagnose.Pod, "test", "readiness-probe-error-6cb7664768-qlmns"),
	Entry("should detect from replicaset", diagnose.ReplicaSet, "test", "readiness-probe-error-6cb7664768"),
	Entry("should detect from service", diagnose.Service, "test", "readiness-probe-error"),
	Entry("should detect from route", diagnose.Route, "test", "readiness-probe-error"),
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
		Expect(logger.Output()).To(ContainSubstring(`👀 checking ReplicaSet 'sa-notfound-59b5d8468f'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 replicaset 'sa-notfound-59b5d8468f' failed to create pods: pods "sa-notfound-59b5d8468f-" is forbidden: error looking up service account test/sa-notfound: serviceaccount "sa-notfound" not found`))
	},
	Entry("should detect from replicaset", diagnose.ReplicaSet, "test", "sa-notfound-59b5d8468f"),
	Entry("should detect from service", diagnose.Service, "test", "sa-notfound"),
	Entry("should detect from route", diagnose.Route, "test", "sa-notfound"),
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
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'no-matching-pods' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no pods matching label selector 'app=invalid' found in namespace 'test'`))
		Expect(logger.Output()).To(ContainSubstring(`💡 you may want to verify that the pods exist and their labels match 'app=invalid'`))
	},
	Entry("should detect from route", diagnose.Route, "test", "no-matching-pods"),
	Entry("should detect from service", diagnose.Service, "test", "no-matching-pods"),
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
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'invalid-service-target-port-str' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no container with port 'https' in pod 'invalid-service-target-port-str-76d5db5c9b-s8wpq'`))
	},
	Entry("should detect from route", diagnose.Route, "test", "invalid-service-target-port-str"),
	Entry("should detect from service", diagnose.Service, "test", "invalid-service-target-port-str"),
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
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'invalid-service-target-port-int' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no container with port '8443' in pod 'invalid-service-target-port-int-bbcb4fd5d-k8kg8'`))
	},
	Entry("should detect from route", diagnose.Route, "test", "invalid-service-target-port-int"),
	Entry("should detect from service", diagnose.Service, "test", "invalid-service-target-port-int"),
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
		Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'unknown-target-service' in namespace 'test'`))
		Expect(logger.Output()).To(ContainSubstring("👻 unable to find service 'unknown'"))
	},
	Entry("should detect from route", diagnose.Route, "test", "unknown-target-service"),
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
		Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'invalid-route-target-port-str' in namespace 'test'`))
		Expect(logger.Output()).To(ContainSubstring("👻 route target port 'https' is not defined in service 'invalid-route-target-port-str'"))
	},
	Entry("should detect from route", diagnose.Route, "test", "invalid-route-target-port-str"),
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
		Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'invalid-route-target-port-int' in namespace 'test'`))
		Expect(logger.Output()).To(ContainSubstring("👻 route target port '8443' is not defined in service 'invalid-route-target-port-int'"))
	},
	Entry("should detect from route", diagnose.Route, "test", "invalid-route-target-port-int"),
)
