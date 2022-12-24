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

func diagnoseDeployment(logger logr.Logger, cfg *rest.Config, namespace, name string) (bool, error) {
	cl := kubernetes.NewForConfigOrDie(cfg)
	d, err := cl.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return checkDeployment(logger, cl, d)
}

func checkDeployment(logger logr.Logger, cl *kubernetes.Clientset, d *appsv1.Deployment) (bool, error) {
	logger.Infof("👀 checking deployment '%s' in namespace '%s'...", d.Name, d.Namespace)
	found := false
	for _, c := range d.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable && c.Status == corev1.ConditionFalse {
			logger.Errorf("👻 %s", c.Message)
			found = true
		}
	}
	// check the `.spec.replicas`
	if d.Spec.Replicas != nil && *d.Spec.Replicas == 0 {
		logger.Errorf("👻 number of desired replicas for deployment '%s' is set to 0", d.Name)
		logger.Infof("💡 run 'oc scale --replicas=1 deployment/%s -n %s' or increase the 'replicas' value in the deployment specs", d.Name, d.Namespace)
		// no need to check further (and avoid infinite loops if coming from service->replicaset->deployment)
		return true, nil
	}
	// check the associated replicasets
	f, err := checkReplicaSets(logger, cl, d.Namespace, d.Spec.Selector.MatchLabels, d.UID)
	return found || f, err
}
