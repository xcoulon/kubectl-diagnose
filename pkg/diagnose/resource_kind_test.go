package diagnose_test

import (
	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("route kind",
	func(kind string, expected bool) {
		Expect(diagnose.IsRoute(kind)).To(Equal(expected))
	},
	Entry("route", "route", true),
	Entry("routes", "routes", true),
	Entry("route.route.openshift.io", "route.route.openshift.io", true),
	Entry("other", "other", false),
)

var _ = DescribeTable("service kind",
	func(kind string, expected bool) {
		Expect(diagnose.IsService(kind)).To(Equal(expected))
	},
	Entry("service", "service", true),
	Entry("services", "services", true),
	Entry("svc", "svc", true),
	Entry("other", "other", false),
)

var _ = DescribeTable("replicaset kind",
	func(kind string, expected bool) {
		Expect(diagnose.IsReplicaSet(kind)).To(Equal(expected))
	},
	Entry("replicaset", "replicaset", true),
	Entry("replicasets", "replicasets", true),
	Entry("replicaset", "rs", true),
	Entry("replicaset.apps", "replicaset.apps", true),
	Entry("other", "other", false),
)

var _ = DescribeTable("deployment kind",
	func(kind string, expected bool) {
		Expect(diagnose.IsDeployment(kind)).To(Equal(expected))
	},
	Entry("replicaset", "deployment", true),
	Entry("deployments", "deployments", true),
	Entry("deploy", "deploy", true),
	Entry("deployment.apps", "deployment.apps", true),
	Entry("other", "other", false),
)

var _ = DescribeTable("pod kind",
	func(kind string, expected bool) {
		Expect(diagnose.IsPod(kind)).To(Equal(expected))
	},
	Entry("pod", "pod", true),
	Entry("pods", "pods", true),
	Entry("po", "po", true),
	Entry("other", "other", false),
)
