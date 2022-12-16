package diagnose

import (
	"context"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getDeployment(cfg *rest.Config, namespace, name string) (*appsv1.Deployment, error) {
	cl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cl.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})

}

func diagnoseDeployment(logger logr.Logger, cfg *rest.Config, d *appsv1.Deployment) (bool, error) {
	logger.Infof("👀 checking deployment '%s' in namespace '%s'...", d.Name, d.Namespace)
	found := false
	for _, c := range d.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable && c.Status == corev1.ConditionFalse {
			logger.Errorf("👻 %s", c.Message)
			found = true
		}
	}
	cl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return false, err
	}
	// check the `.spec.replicas`
	if d.Spec.Replicas != nil && *d.Spec.Replicas == 0 {
		logger.Errorf("👻 number of desired replicas for deployment '%s' is set to 0", d.Name)
		logger.Infof("💡 run 'oc scale --replicas=1 deployment/%s -n %s' or increase the 'replicas' value in the deployment specs", d.Name, d.Namespace)
		// no need to check further (and avoid infinite loops if coming from service->replicaset->deployment)
		return true, nil
	}
	// check the associated replicasets
	selector := labels.Set(d.Spec.Selector.MatchLabels).String()
	rss, err := cl.AppsV1().ReplicaSets(d.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return false, err
	}
	for i := range rss.Items {
		rs := rss.Items[i]
		for _, ref := range rs.OwnerReferences {
			if ref.UID == d.UID {
				// rs is "owned" by this deployment
				f, err := diagnoseReplicaSet(logger, cfg, &rs)
				if err != nil {
					return false, err
				}
				found = found || f
			}
		}
	}
	return found, nil
}
