package diagnose

import (
	"context"
	"strconv"

	"github.com/charmbracelet/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func diagnoseReplicaSet(ctx context.Context, logger *log.Logger, cfg *rest.Config, namespace, name string) (bool, error) {
	cl := kubernetes.NewForConfigOrDie(cfg)
	rs, err := cl.AppsV1().ReplicaSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return checkReplicaSet(ctx, logger, cl, rs)
}

func checkReplicaSets(ctx context.Context, logger *log.Logger, cl *kubernetes.Clientset, namespace string, selector map[string]string, ownerID types.UID) (bool, error) {
	rss, err := cl.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(selector).String(),
	})
	if err != nil {
		return false, err
	}

	var rs *appsv1.ReplicaSet
rss:
	for i := range rss.Items {
		for _, ref := range rss.Items[i].OwnerReferences {
			if ref.UID == ownerID {
				rsRevision := 0
				if rs != nil {
					rsRevision, _ = strconv.Atoi(rs.Annotations["deployment.kubernetes.io/revision"])
				}
				itemRevision, _ := strconv.Atoi(rss.Items[i].Annotations["deployment.kubernetes.io/revision"])
				if itemRevision > rsRevision {
					rs = &rss.Items[i]
					continue rss
				}
			}
		}
	}
	// only check the latest ReplicaSet (older ones may be associated with previous configs, and should eventually be deleted once pods are running)
	return checkReplicaSet(ctx, logger, cl, rs)

}

func checkReplicaSet(ctx context.Context, logger *log.Logger, cl *kubernetes.Clientset, rs *appsv1.ReplicaSet) (bool, error) {
	logger.Infof("👀 checking replicaset '%s' in namespace '%s'...", rs.Name, rs.Namespace)
	for _, c := range rs.Status.Conditions {
		if c.Type == appsv1.ReplicaSetReplicaFailure &&
			c.Reason == "FailedCreate" &&
			c.Status == corev1.ConditionTrue {
			logger.Errorf("👻 replicaset '%s' failed to create pods: %s", rs.Name, c.Message)
			return true, nil
		}
	}
	// check the `.spec.replicas`
	if rs.Spec.Replicas != nil && *rs.Spec.Replicas == 0 {
		for _, ownerRef := range rs.OwnerReferences {
			if ownerRef.Kind == "Deployment" {
				d, err := cl.AppsV1().Deployments(rs.Namespace).Get(ctx, ownerRef.Name, metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				return checkDeployment(ctx, logger, cl, d)
			}
		}
	}
	// if status looks fine, then look for pods with the matching label(s)
	selector := labels.Set(rs.Spec.Selector.MatchLabels).String()
	pods, err := cl.CoreV1().Pods(rs.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return false, err
	}
	// if there is no pod matching the selector
	if len(pods.Items) == 0 {
		logger.Errorf("👻 no pods matching label selector '%s' found in namespace '%s'", selector, rs.Namespace)
		logger.Infof("💡 you may want to verify that the pods exist and their labels match '%s'", selector)
		return true, nil
	}
	for i := range pods.Items {
		pod := pods.Items[i]
		logger.Infof("👀 checking pod '%s'...", pod.Name)
		for _, ownerRef := range pod.OwnerReferences {
			if ownerRef.UID == rs.UID {
				// pod is "owned" by this replicaset
				if found, err := checkPod(ctx, logger, cl, &pod); err != nil {
					return false, err
				} else if found {
					return true, nil
				}
			}
		}
	}
	return false, nil
}
