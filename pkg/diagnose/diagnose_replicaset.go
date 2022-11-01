package diagnose

import (
	"context"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getReplicaSet(cfg *rest.Config, namespace, name string) (*appsv1.ReplicaSet, error) {
	cl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cl.AppsV1().ReplicaSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})

}

func checkReplicaSet(logger logr.Logger, rs *appsv1.ReplicaSet) (bool, error) {
	logger.Infof("👀 checking replicaset '%s'...", rs.Name)
	return checkReplicaSetStatus(logger, rs)
}

// check the status of the pod status
func checkReplicaSetStatus(logger logr.Logger, rs *appsv1.ReplicaSet) (bool, error) {
	for _, c := range rs.Status.Conditions {
		if c.Type == appsv1.ReplicaSetReplicaFailure &&
			c.Reason == "FailedCreate" &&
			c.Status == corev1.ConditionTrue {
			logger.Errorf("👻 replicaset '%s' failed to create pods: %s", rs.Name, c.Message)
			return true, nil
		}
	}
	return false, nil
}
