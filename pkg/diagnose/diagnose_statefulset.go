package diagnose

import (
	"context"

	"github.com/charmbracelet/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func diagnoseStatefulSet(ctx context.Context, logger *log.Logger, cfg *rest.Config, namespace, name string) (bool, error) {
	cl := kubernetes.NewForConfigOrDie(cfg)
	sts, err := cl.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return checkStatefulSet(ctx, logger, cl, sts)
}

func checkStatefulSet(ctx context.Context, logger *log.Logger, cl *kubernetes.Clientset, sts *appsv1.StatefulSet) (bool, error) {
	logger.Debugf("👀 checking statefulset '%s' in namespace '%s'...", sts.Name, sts.Namespace)
	found := false
	// check the replicas
	if sts.Spec.Replicas != nil && *sts.Spec.Replicas == 0 {
		logger.Errorf("👻 number of desired replicas for statefulset '%s' is set to 0", sts.Name)
		logger.Infof("💡 run 'oc scale --replicas=1 sts/%s -n %s' or increase the 'replicas' value in the statefulset specs", sts.Name, sts.Namespace)
		// no need to check further
		return true, nil
	}
	// check events associated with the statefulset
	if found, err := checkEvents(ctx, logger, cl, sts); found || err != nil {
		return found, err
	}

	logger.Debugf("👀 checking statefulset status...")
	// checking the pods
	// TODO: remove code duplication with ReplicaSet checks
	selector := labels.Set(sts.Spec.Selector.MatchLabels).String()
	pods, err := cl.CoreV1().Pods(sts.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return false, err
	}
	// if there is no pod matching the selector
	if len(pods.Items) == 0 {
		logger.Errorf("👻 no pods matching label selector '%s' found in namespace '%s'", selector, sts.Namespace)
		logger.Infof("💡 you may want to verify that the pods exist and their labels match '%s'", selector)
		return true, nil
	}
	for i := range pods.Items {
		pod := pods.Items[i]
		logger.Debugf("👀 checking pod '%s'...", pod.Name)
		for _, ownerRef := range pod.OwnerReferences {
			if ownerRef.UID == sts.UID {
				// pod is "owned" by this sts
				if found, err := checkPod(ctx, logger, cl, &pod); err != nil {
					return false, err
				} else if found {
					return true, nil
				}
			}
		}
	}
	return found, nil
}
