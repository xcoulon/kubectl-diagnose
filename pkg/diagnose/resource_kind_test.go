package diagnose_test

import (
	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("resource kind",
	func(kind string, expected int) {
		Expect(diagnose.Kind(kind)).To(Equal(expected))
	},
	// routes
	Entry("routes", "routes", diagnose.Route),
	Entry("route", "route", diagnose.Route),
	Entry("route.route.openshift.io", "route.route.openshift.io", diagnose.Route),

	// services
	Entry("services", "services", diagnose.Service),
	Entry("service", "service", diagnose.Service),
	Entry("svc", "svc", diagnose.Service),

	// replicasets
	Entry("replicasets", "replicasets", diagnose.ReplicaSet),
	Entry("replicaset", "replicaset", diagnose.ReplicaSet),
	Entry("rs", "rs", diagnose.ReplicaSet),
	Entry("replicaset.apps", "replicaset.apps", diagnose.ReplicaSet),

	// deployments
	Entry("deployments", "deployments", diagnose.Deployment),
	Entry("deployment", "deployment", diagnose.Deployment),
	Entry("deploy", "deploy", diagnose.Deployment),
	Entry("deployment.apps", "deployment.apps", diagnose.Deployment),

	// pods
	Entry("pods", "pods", diagnose.Pod),
	Entry("pod", "pod", diagnose.Pod),
	Entry("po", "po", diagnose.Pod),

	// statefulsets
	Entry("statefulsets", "statefulsets", diagnose.StatefulSet),
	Entry("statefulset", "statefulset", diagnose.StatefulSet),
	Entry("sts", "sts", diagnose.StatefulSet),
	Entry("statefulset.apps", "statefulset.apps", diagnose.StatefulSet),

	// persistent volume claims
	Entry("persistentvolumeclaims", "persistentvolumeclaims", diagnose.PersistentVolumeClaim),
	Entry("persistentvolumeclaim", "persistentvolumeclaim", diagnose.PersistentVolumeClaim),
	Entry("pvc", "pvc", diagnose.PersistentVolumeClaim),

	// storage classe
	Entry("storageclass", "storageclass", diagnose.StorageClass),
	Entry("storageclasses", "storageclasses", diagnose.StorageClass),
	Entry("sc", "sc", diagnose.StorageClass),
	Entry("storageclass.storage.k8s.io", "storageclass.storage.k8s.io", diagnose.StorageClass),
	// unknown
	Entry("other", "other", diagnose.Unkwown),
)
