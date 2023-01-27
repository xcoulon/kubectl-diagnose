package diagnose

import (
	"context"
	"strings"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func diagnosePod(ctx context.Context, logger logr.Logger, cfg *rest.Config, namespace, name string) (bool, error) {
	cl := kubernetes.NewForConfigOrDie(cfg)
	pod, err := cl.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return checkPod(ctx, logger, cl, pod)
}

func findPods(ctx context.Context, cl *kubernetes.Clientset, namespace string, selector map[string]string) ([]corev1.Pod, error) {
	// find all pods with the associated label selector in the same namespace
	pods, err := cl.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(selector).String(),
	})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func checkPod(ctx context.Context, logger logr.Logger, cl *kubernetes.Clientset, pod *corev1.Pod) (bool, error) {
	logger.Infof("👀 checking pod '%s' in namespace '%s'...", pod.Name, pod.Namespace)
	found := false
	logger.Debugf("👀 checking pod status...")
	//
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.ContainersReady && c.Status == corev1.ConditionTrue {
			logger.Infof("☑️  all containers in pod '%s' are ready", pod.Name)
			return false, nil
		}
	}
	// check the containers
	for _, c := range pod.Status.Conditions {
		switch {
		case c.Type == corev1.ContainersReady && c.Status == corev1.ConditionFalse:
			if c.Message != "" {
				logger.Errorf("👻 %s", c.Message)
			}
			// also, check the container statuses
			f, err := diagnoseContainer(ctx, logger, cl, pod)
			if err != nil {
				return false, err
			}
			found = found || f
		case c.Type == corev1.PodScheduled && c.Status == corev1.ConditionFalse && c.Reason == corev1.PodReasonUnschedulable:
			// check if there's a pending PVC
			for _, v := range pod.Spec.Volumes {
				if v.PersistentVolumeClaim != nil {
					pvc, err := cl.CoreV1().PersistentVolumeClaims(pod.Namespace).Get(ctx, v.PersistentVolumeClaim.ClaimName, metav1.GetOptions{})
					if err != nil {
						return false, err
					}
					f, err := checkPersistentVolumeClaim(ctx, logger, cl, pvc)
					if err != nil {
						return false, err
					}
					found = found || f
				}
			}
		}
	}

	// check events associated with the pod
	f, err := checkEvents(ctx, logger, cl, pod)
	if err != nil {
		return false, err
	}
	found = found || f
	return found, nil
}

// check the status of the pod containers
// return the list of containers' name whose status is `waiting`
func diagnoseContainer(ctx context.Context, logger logr.Logger, cl *kubernetes.Clientset, pod *corev1.Pod) (bool, error) {
	found := false
	for _, cs := range pod.Status.ContainerStatuses {
		switch {
		case cs.State.Running != nil && cs.Started != nil && !*cs.Started: // container is still starting...
			found = true
			logger.Errorf("👻 container '%s' is still starting", cs.Name)
		case cs.State.Waiting != nil: // container is waiting
			found = true
			if cs.State.Waiting.Message != "" {
				logger.Errorf("👻 container '%s' is waiting with reason '%s': %s", cs.Name, cs.State.Waiting.Reason, cs.State.Waiting.Message)
			} else {
				logger.Errorf("👻 container '%s' is waiting with reason '%s'", cs.Name, cs.State.Waiting.Reason)
			}
			switch {
			case cs.State.Waiting.Reason == "CrashLoopBackOff" && cs.LastTerminationState.Terminated != nil && cs.LastTerminationState.Terminated.Message != "":
				logger.Errorf("🗒 %s", strings.ReplaceAll(cs.LastTerminationState.Terminated.Message, "\n", "\n  "))
			case cs.State.Waiting != nil && cs.State.Waiting.Reason == "ContainerCreating":
				// do nothing
			default:
				_, err := checkContainerLogs(ctx, logger, cl, pod, cs.Name)
				if err != nil {
					return false, err
				}
			}

		case cs.Started != nil && *cs.Started && !cs.Ready:
			f, err := checkContainerLogs(ctx, logger, cl, pod, cs.Name)
			if err != nil {
				return false, err
			}
			found = found || f
		}
		// also, check the logs
		// if () ||
		// 	s.LastTerminationState.Running != nil ||
		// 	s.LastTerminationState.Terminated != nil ||
		// 	s.LastTerminationState.Waiting != nil {
		// }
		// TODO: check reason and provide a more detailed diagnosis or hint to fix the problem?
		// eg: if reason is `CrashLoopBackOff`, look for errors (`ERROR`/`FATAL`) in the container logs? (but display the n last lines?)
		// eg: if reason is `CreateContainerConfigError`, message should be enough (eg: `secret "cookie" not found`)
	}
	return found, nil
}

func checkContainerLogs(ctx context.Context, logger logr.Logger, cl *kubernetes.Clientset, pod *corev1.Pod, container string) (bool, error) {
	found := false
	logger.Debugf("👀 checking '%s' container logs...", container)
	logs, err := cl.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: container}).DoRaw(ctx)
	if err != nil {
		return false, err
	}
	logger.Debugf("logs: '%s'", string(logs))
	for _, l := range strings.Split(string(logs), "\n") {
		ll := strings.ToLower(l)
		if strings.Contains(ll, "error") ||
			strings.Contains(ll, "failed") ||
			strings.Contains(ll, "fatal") ||
			strings.Contains(ll, "panic") ||
			strings.Contains(ll, "emerg") {
			found = true
			logger.Errorf("🗒  %s", l)
		}
	}
	if !found {
		logger.Infof("🤷 no 'error'/'failed'/'fatal'/'panic'/'emerg' messages found in the '%s' container logs", container)
	}
	return found, nil
}
