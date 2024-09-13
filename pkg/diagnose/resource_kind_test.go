package diagnose_test

import (
	"testing"

	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"

	"github.com/stretchr/testify/assert"
)

func TestResourceKind(t *testing.T) {

	testcases := []struct {
		name     string
		expected diagnose.ResourceKind
	}{
		// routes
		{
			name:     "routes",
			expected: diagnose.Route,
		},
		{
			name:     "route",
			expected: diagnose.Route,
		},
		{
			name:     "route.route.openshift.io",
			expected: diagnose.Route,
		},

		// services
		{
			name:     "services",
			expected: diagnose.Service,
		},
		{
			name:     "service",
			expected: diagnose.Service,
		},
		{
			name:     "svc",
			expected: diagnose.Service,
		},

		// replicasets
		{
			name:     "replicasets",
			expected: diagnose.ReplicaSet,
		},
		{
			name:     "replicaset",
			expected: diagnose.ReplicaSet,
		},
		{
			name:     "rs",
			expected: diagnose.ReplicaSet,
		},
		{
			name:     "replicaset.apps",
			expected: diagnose.ReplicaSet,
		},

		// deployments
		{
			name:     "deployments",
			expected: diagnose.Deployment,
		},
		{
			name:     "deployment",
			expected: diagnose.Deployment,
		},
		{
			name:     "deploy",
			expected: diagnose.Deployment,
		},
		{
			name:     "deployment.apps",
			expected: diagnose.Deployment,
		},

		// pods
		{
			name:     "pods",
			expected: diagnose.Pod,
		},
		{
			name:     "pod",
			expected: diagnose.Pod,
		},
		{
			name:     "po",
			expected: diagnose.Pod,
		},

		// statefulsets
		{
			name:     "statefulsets",
			expected: diagnose.StatefulSet,
		},
		{
			name:     "statefulset",
			expected: diagnose.StatefulSet,
		},
		{
			name:     "sts",
			expected: diagnose.StatefulSet,
		},
		{
			name:     "statefulset.apps",
			expected: diagnose.StatefulSet,
		},

		// persistent volume claims
		{
			name:     "persistentvolumeclaims",
			expected: diagnose.PersistentVolumeClaim,
		},
		{
			name:     "persistentvolumeclaim",
			expected: diagnose.PersistentVolumeClaim,
		},
		{
			name:     "pvc",
			expected: diagnose.PersistentVolumeClaim,
		},

		// storage classes
		{
			name:     "storageclasses",
			expected: diagnose.StorageClass,
		},
		{
			name:     "storageclass",
			expected: diagnose.StorageClass,
		},
		{
			name:     "sc",
			expected: diagnose.StorageClass,
		},
		{
			name:     "storageclass.storage.k8s.io",
			expected: diagnose.StorageClass,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			actual := diagnose.NewResourceKind(tc.name)
			// then
			assert.Equal(t, tc.expected, actual)
		})
	}
}
