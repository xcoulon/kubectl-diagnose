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

func checkDeployment(logger logr.Logger, cfg *rest.Config, d *appsv1.Deployment) (bool, error) {
	logger.Infof("👀 checking deployment '%s'...", d.Name)
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
				f, err := checkReplicaSet(logger, cfg, &rs)
				if err != nil {
					return false, err
				}
				found = found || f
			}
		}
	}
	return found, nil
}
