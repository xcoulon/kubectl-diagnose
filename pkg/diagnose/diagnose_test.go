package diagnose_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	"github.com/xcoulon/kubectl-diagnose/testsupport"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// ----------------------------
// Routes
// ----------------------------
var _ = DescribeTable("should detect missing route target service",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/route-unknown-target-service.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'unknown-target-service' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring("👻 unable to find service 'unknown'"))
	},
	Entry("from route", diagnose.Route, "test", "unknown-target-service"),
)

var _ = DescribeTable("should detect invalid route target port as string",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/route-invalid-target-port-str.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'invalid-route-target-port-str' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring("👻 route target port 'https' is not defined in service 'invalid-route-target-port-str'"))
	},
	Entry("from route", diagnose.Route, "test", "invalid-route-target-port-str"),
)

var _ = DescribeTable("should detect invalid route target port as int",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/route-invalid-target-port-int.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'invalid-route-target-port-int' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring("👻 route target port '8443' is not defined in service 'invalid-route-target-port-int'"))
	},
	Entry("from route", diagnose.Route, "test", "invalid-route-target-port-int"),
)

// ----------------------------
// Services
// ----------------------------

var _ = DescribeTable("should detect no matching pods",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/service-no-matching-pods.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		if kind == diagnose.Route {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'no-matching-pods' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'no-matching-pods' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no pods matching label selector 'app=invalid' found in namespace 'test'`))
		Expect(logger.Output()).To(ContainSubstring(`💡 you may want to verify that the pods exist and their labels match 'app=invalid'`))
	},
	Entry("from service", diagnose.Service, "test", "no-matching-pods"),
	Entry("from route", diagnose.Route, "test", "no-matching-pods"),
)

var _ = DescribeTable("should detect invalid service target port as string",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/service-invalid-target-port-str.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		if kind == diagnose.Route {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'invalid-service-target-port-str' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'invalid-service-target-port-str' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no container with port 'https' in pod 'invalid-service-target-port-str-76d5db5c9b-s8wpq'`))
	},
	Entry("from service", diagnose.Service, "test", "invalid-service-target-port-str"),
	Entry("from route", diagnose.Route, "test", "invalid-service-target-port-str"),
)

var _ = DescribeTable("should detect invalid service target port as int",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/service-invalid-target-port-int.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		if kind == diagnose.Route {
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'invalid-service-target-port-int' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'invalid-service-target-port-int' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no container with port '8443' in pod 'invalid-service-target-port-int-bbcb4fd5d-k8kg8'`))
	},
	Entry("from service", diagnose.Service, "test", "invalid-service-target-port-int"),
	Entry("from route", diagnose.Route, "test", "invalid-service-target-port-int"),
)

// ----------------------------
// Deployments / ReplicaSets
// ----------------------------
var _ = DescribeTable("should detect zero replicas specified in deployment",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/deployment-zero-replica.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		switch {
		case kind == diagnose.Route:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'deploy-zero-replica' in namespace 'test'...`))
		case kind == diagnose.Route || kind == diagnose.Service:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'deploy-zero-replica' in namespace 'test'...`))
		case kind == diagnose.Route || kind == diagnose.Service || kind == diagnose.ReplicaSet:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'deploy-zero-replica-9bccf7d88' in namespace 'test'...`))
		}
		// in all cases
		Expect(logger.Output()).To(ContainSubstring(`👀 checking deployment 'deploy-zero-replica' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 number of desired replicas for deployment 'deploy-zero-replica' is set to 0`))
		Expect(logger.Output()).To(ContainSubstring(`💡 run 'oc scale --replicas=1 deployment/deploy-zero-replica -n test' or increase the 'replicas' value in the deployment specs`))
		Expect(logger.Output()).NotTo(ContainSubstring("👻 no pods matching label selector")) // should not appear, other messages are enough
	},
	Entry("from replicaset", diagnose.ReplicaSet, "test", "deploy-zero-replica-9bccf7d88"),
	Entry("from deployment", diagnose.Deployment, "test", "deploy-zero-replica"),
	Entry("from service", diagnose.Service, "test", "deploy-zero-replica"),
	Entry("from route", diagnose.Route, "test", "deploy-zero-replica"),
)

var _ = DescribeTable("should detect invalid serviceaccount specified in deployment",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/deployment-service-account-not-found.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		switch {
		case kind == diagnose.Route:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'deploy-sa-notfound' in namespace 'test'...`))
		case kind == diagnose.Route || kind == diagnose.Service:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'deploy-sa-notfound' in namespace 'test'...`))
		case kind == diagnose.Deployment:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking deployment 'deploy-sa-notfound' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'deploy-sa-notfound-59b5d8468f' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 replicaset 'deploy-sa-notfound-59b5d8468f' failed to create pods: pods "deploy-sa-notfound-59b5d8468f-" is forbidden: error looking up service account test/deploy-sa-notfound: serviceaccount "deploy-sa-notfound" not found`))
		Expect(logger.Output()).NotTo(ContainSubstring("👻 no pods matching label selector")) // should not appear, other messages are enough
	},
	Entry("from replicaset", diagnose.ReplicaSet, "test", "deploy-sa-notfound-59b5d8468f"),
	Entry("from deployment", diagnose.Deployment, "test", "deploy-sa-notfound"),
	Entry("from service", diagnose.Service, "test", "deploy-sa-notfound"),
	Entry("from route", diagnose.Route, "test", "deploy-sa-notfound"),
)

// ----------------------------
// StatefulSets
// ----------------------------

var _ = DescribeTable("should detect zero replicas specified in deployment",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/statefulset-zero-replica.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		switch {
		case kind == diagnose.Route:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'sts-zero-replica' in namespace 'test'...`))
		case kind == diagnose.Route || kind == diagnose.Service:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'sts-zero-replica' in namespace 'test'...`))
		case kind == diagnose.StatefulSet:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking statefulset 'sts-zero-replica' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👻 number of desired replicas for statefulset 'sts-zero-replica' is set to 0`))
		Expect(logger.Output()).To(ContainSubstring(`💡 run 'oc scale --replicas=1 sts/sts-zero-replica -n test' or increase the 'replicas' value in the statefulset specs`))
		Expect(logger.Output()).NotTo(ContainSubstring("👻 no pods matching label selector")) // should not appear, other messages are enough

	},
	Entry("from statefulset", diagnose.StatefulSet, "test", "sts-zero-replica"),
	Entry("from service", diagnose.Service, "test", "sts-zero-replica"),
	Entry("from route", diagnose.Route, "test", "sts-zero-replica"),
)

var _ = DescribeTable("should detect invalid serviceaccount specified in statefulset",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/statefulset-service-account-not-found.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		switch {
		case kind == diagnose.Route:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'sts-sa-notfound' in namespace 'test'...`))
		case kind == diagnose.Route || kind == diagnose.Service:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'sts-sa-notfound' in namespace 'test'...`))
		case kind == diagnose.StatefulSet:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking statefulset 'sts-sa-notfound' in namespace 'test'...`))
		}
		// in all cases:
		// events
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-27T08:51:34Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: FailedCreate: create Pod sts-sa-notfound-0 in StatefulSet sts-sa-notfound failed error: pods "sts-sa-notfound-0" is forbidden: error looking up service account test/unknown: serviceaccount "unknown" not found`, time.Since(lastTimestamp).Truncate(time.Second))))
		Expect(logger.Output()).NotTo(ContainSubstring("👻 no pods matching label selector")) // should not appear, other messages are enough
	},
	Entry("from statefulset", diagnose.StatefulSet, "test", "sts-sa-notfound"),
	Entry("from service", diagnose.Service, "test", "sts-sa-notfound"),
	Entry("from route", diagnose.Route, "test", "sts-sa-notfound"),
)

var _ = DescribeTable("should detect invalid storageclass specified in statefulset",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/statefulset-invalid-storageclass.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		switch {
		case kind == diagnose.Route:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'sts-invalid-sc' in namespace 'test'...`))
		case kind == diagnose.Route || kind == diagnose.Service:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'sts-invalid-sc' in namespace 'test'...`))
		case kind == diagnose.StatefulSet:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking statefulset 'sts-invalid-sc' in namespace 'test'...`))
		}
		// in all cases:
		// pod events
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-26T08:40:16.475828Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: FailedScheduling: 0/12 nodes are available: 12 pod has unbound immediate PersistentVolumeClaims. preemption: 0/12 nodes are available: 12 Preemption is not helpful for scheduling.`, time.Since(lastTimestamp).Truncate(time.Second))))
		// associated persistent volume claim
		lastTimestamp, _ = time.Parse("2006-01-02T15:04:05Z", "2022-11-26T09:40:20Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: ProvisioningFailed: storageclass.storage.k8s.io "unknown" not found`, time.Since(lastTimestamp).Truncate(time.Second))))

	},
	Entry("from pod", diagnose.Pod, "test", "sts-invalid-sc-0"),
	Entry("from statefulset", diagnose.StatefulSet, "test", "sts-invalid-sc"),
	Entry("from service", diagnose.Service, "test", "sts-invalid-sc"),
	Entry("from route", diagnose.Route, "test", "sts-invalid-sc"),
)

// ----------------------------
// Pods
// ----------------------------

var _ = DescribeTable("should detect deployment pod container in CrashLoopBackOff status",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/deployment-pod-crash-loop-back-off.yaml", "resources/deployment-pod-crash-loop-back-off.logs")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		switch {
		case kind == diagnose.Route:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'deploy-crash-loop-back-off' in namespace 'test'...`))
		case kind == diagnose.Route || kind == diagnose.Service:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'deploy-crash-loop-back-off' in namespace 'test'...`))
		case kind == diagnose.Deployment:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking deployment 'deploy-crash-loop-back-off' in namespace 'test'...`))
		case kind == diagnose.Deployment || kind == diagnose.ReplicaSet:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'deploy-crash-loop-back-off-7994787459' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'deploy-crash-loop-back-off-7994787459-2nrz5' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		// event
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-12T18:02:28Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: BackOff: Back-off restarting failed container`, time.Since(lastTimestamp).Truncate(time.Second))))
		// logs
		Expect(logger.Output()).To(ContainSubstring(`🗒  Error: loading initial config: loading new config: http app module: start: listening on :80: listen tcp :80: bind: permission denied`))
	},
	Entry("from pod", diagnose.Pod, "test", "deploy-crash-loop-back-off-7994787459-2nrz5"),
	Entry("from replicaset", diagnose.ReplicaSet, "test", "deploy-crash-loop-back-off-7994787459"),
	Entry("from deployment", diagnose.Deployment, "test", "deploy-crash-loop-back-off"),
	Entry("from service", diagnose.Service, "test", "deploy-crash-loop-back-off"),
	Entry("from route", diagnose.Route, "test", "deploy-crash-loop-back-off"),
)

var _ = DescribeTable("should detect deployment pod container in ImagePullBackOff status",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/deployment-pod-image-pull-back-off.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		switch {
		case kind == diagnose.Route:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'deploy-image-pull-back-off' in namespace 'test'...`))
		case kind == diagnose.Route || kind == diagnose.Service:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'deploy-image-pull-back-off' in namespace 'test'...`))
		case kind == diagnose.Deployment:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking deployment 'deploy-image-pull-back-off' in namespace 'test'...`))
		case kind == diagnose.Deployment || kind == diagnose.ReplicaSet:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'deploy-image-pull-back-off-9bbb4f9bd' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'ImagePullBackOff': Back-off pulling image "unknown:v0.0.0"`))
		// events
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-13T07:59:04Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Failed: Error: ImagePullBackOff`, time.Since(lastTimestamp).Truncate(time.Second))))
	},
	Entry("from pod", diagnose.Pod, "test", "deploy-image-pull-back-off-9bbb4f9bd-pjj55"),
	Entry("from replicaset", diagnose.ReplicaSet, "test", "deploy-image-pull-back-off-9bbb4f9bd"),
	Entry("from deployment", diagnose.Deployment, "test", "deploy-image-pull-back-off"),
	Entry("from service", diagnose.Service, "test", "deploy-image-pull-back-off"),
	Entry("from route", diagnose.Route, "test", "deploy-image-pull-back-off"),
)

var _ = DescribeTable("should detect deployment pod container with readiness probe error",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/deployment-pod-readiness-probe-error.yaml", "resources/deployment-pod-readiness-probe-error.logs")
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		switch {
		case kind == diagnose.Route:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'deploy-readiness-probe-error' in namespace 'test'...`))
		case kind == diagnose.Route || kind == diagnose.Service:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'deploy-readiness-probe-error' in namespace 'test'...`))
		case kind == diagnose.Deployment:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking deployment 'deploy-readiness-probe-error' in namespace 'test'...`))
		case kind == diagnose.Deployment || kind == diagnose.ReplicaSet:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'deploy-readiness-probe-error-6cb7664768' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'deploy-readiness-probe-error-6cb7664768-qlmns' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		// events
		// events
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-13T21:55:27Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Unhealthy: Readiness probe failed: HTTP probe failed with statuscode: 404`, time.Since(lastTimestamp).Truncate(time.Second))))
		// logs
		Expect(logger.Output()).To(ContainSubstring("🤷 no 'error'/'fatal'/'panic'/'emerg' messages found in the container logs"))
	},
	Entry("from pod", diagnose.Pod, "test", "deploy-readiness-probe-error-6cb7664768-qlmns"),
	Entry("from replicaset", diagnose.ReplicaSet, "test", "deploy-readiness-probe-error-6cb7664768"),
	Entry("from deployment", diagnose.Deployment, "test", "deploy-readiness-probe-error"),
	Entry("from service", diagnose.Service, "test", "deploy-readiness-probe-error"),
	Entry("from route", diagnose.Route, "test", "deploy-readiness-probe-error"),
)

var _ = DescribeTable("should detect deployment pod container with unknown configmap mount",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/deployment-pod-unknown-configmap.yaml") // no logs, container is not created
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		switch {
		case kind == diagnose.Route:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'deploy-unknown-cm' in namespace 'test'...`))
		case kind == diagnose.Route || kind == diagnose.Service:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'deploy-unknown-cm' in namespace 'test'...`))
		case kind == diagnose.Deployment:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking deployment 'deploy-unknown-cm' in namespace 'test'...`))
		case kind == diagnose.Deployment || kind == diagnose.ReplicaSet:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'deploy-unknown-cm-76476b7d5' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'deploy-unknown-cm-76476b7d5-q2khp' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'ContainerCreating'`))
		Expect(logger.Output()).NotTo(ContainSubstring(`👻 container 'default' is waiting with reason 'ContainerCreating':`)) // ensure there is no `:` followed by an empty message
		// no logs: container has not started
		// events
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-11-13T17:19:34Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: FailedMount: Unable to attach or mount volumes: unmounted volumes=[caddy-config], unattached volumes=[caddy-config caddy-config-cache kube-api-access-62xrc]: timed out waiting for the condition`, time.Since(lastTimestamp).Truncate(time.Second))))
	},
	Entry("from pod", diagnose.Pod, "test", "deploy-unknown-cm-76476b7d5-q2khp"),
	Entry("from replicaset", diagnose.ReplicaSet, "test", "deploy-unknown-cm-76476b7d5"),
	Entry("from deployment", diagnose.Deployment, "test", "deploy-unknown-cm"),
	Entry("from service", diagnose.Service, "test", "deploy-unknown-cm"),
	Entry("from route", diagnose.Route, "test", "deploy-unknown-cm"),
)

var _ = DescribeTable("should detect statefulset pod container with unknown configmap mount",
	func(kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger, "resources/statefulset-pod-unknown-configmap.yaml") // no logs, container is not created
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		switch {
		case kind == diagnose.Route:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'sts-unknown-cm' in namespace 'test'...`))
		case kind == diagnose.Route || kind == diagnose.Service:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'sts-unknown-cm' in namespace 'test'...`))
		case kind == diagnose.StatefulSet:
			Expect(logger.Output()).To(ContainSubstring(`👀 checking statefulset 'sts-unknown-cm' in namespace 'test'...`))
		}
		// in all cases:
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'sts-unknown-cm-0' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'CreateContainerConfigError': configmap "sts-unknown-cm" not found`))
		// no logs: container has not started
		// events
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-12-01T05:40:55Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Failed: Error: configmap "sts-unknown-cm" not found`, time.Since(lastTimestamp).Truncate(time.Second))))
	},
	Entry("from statefulset", diagnose.StatefulSet, "test", "sts-unknown-cm"),
	Entry("from service", diagnose.Service, "test", "sts-unknown-cm"),
	Entry("from route", diagnose.Route, "test", "sts-unknown-cm"),
)

// ----------------------------
// Errors
// ----------------------------

var _ = DescribeTable("should handle internal server errors",
	func(gr string, kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger)
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		_, err = diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(apierrors.IsInternalError(err)).To(BeTrue())
	},
	Entry("from pod", "pods", diagnose.Pod, "test", "error"),
	Entry("from persistentvolumeclaim", "persistentvolumeclaims", diagnose.PersistentVolumeClaim, "test", "error"),
	Entry("from statefulset", "statefulsets.apps", diagnose.StatefulSet, "test", "error"),
	Entry("from deployment", "deployments.apps", diagnose.Deployment, "test", "error"),
	Entry("from service", "services", diagnose.Service, "test", "error"),
	Entry("from route", "routes.route.openshift.io", diagnose.Route, "test", "error"),
)

var _ = DescribeTable("should handle not found errors",
	func(gr string, kind diagnose.ResourceKind, namespace, name string) {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := testsupport.NewFakeAPIServer(logger)
		Expect(err).NotTo(HaveOccurred())
		cfg := testsupport.NewConfig(apiserver.URL, "/api")

		// when
		_, err = diagnose.Diagnose(logger, cfg, kind, namespace, name)

		// then
		Expect(err).To(BeANotFoundError())
	},
	Entry("from pod", "pods", diagnose.Pod, "test", "notfound"),
	Entry("from persistentvolumeclaim", "persistentvolumeclaims", diagnose.PersistentVolumeClaim, "test", "notfound"),
	Entry("from statefulset", "statefulsets.apps", diagnose.StatefulSet, "test", "notfound"),
	Entry("from deployment", "deployments.apps", diagnose.Deployment, "test", "notfound"),
	Entry("from service", "services", diagnose.Service, "test", "notfound"),
	Entry("from route", "routes.route.openshift.io", diagnose.Route, "test", "notfound"),
)

func BeANotFoundError() types.GomegaMatcher {
	return And(
		WithTransform(func(err error) (int, error) {
			if e := apierrors.APIStatus(nil); errors.As(err, &e) {
				return int(e.Status().Code), nil
			}
			return -1, fmt.Errorf("wrong type of error")
		}, Equal(http.StatusNotFound)),
	)
}
