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

func diagnosePod(logger logr.Logger, cfg *rest.Config, namespace, name string) (bool, error) {
	cl := kubernetes.NewForConfigOrDie(cfg)
	pod, err := cl.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return checkPod(logger, cl, pod)
}

func findPods(cl *kubernetes.Clientset, namespace string, selector map[string]string) ([]corev1.Pod, error) {
	// find all pods with the associated label selector in the same namespace
	pods, err := cl.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labels.Set(selector).String(),
	})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func checkPod(logger logr.Logger, cl *kubernetes.Clientset, pod *corev1.Pod) (bool, error) {
	logger.Infof("👀 checking pod '%s' in namespace '%s'...", pod.Name, pod.Namespace)
	found := false
	logger.Debugf("👀 checking pod status...")
	// check the containers
	for _, c := range pod.Status.Conditions {
		switch {
		case c.Type == corev1.ContainersReady && c.Status == corev1.ConditionFalse:
			if c.Message != "" {
				logger.Errorf("👻 %s", c.Message)
			}
			// also, check the container statuses
			f, err := diagnoseContainer(logger, cl, pod)
			if err != nil {
				return false, err
			}
			found = found || f
		case c.Type == corev1.PodScheduled && c.Status == corev1.ConditionFalse && c.Reason == corev1.PodReasonUnschedulable:
			// check if there's a pending PVC
			for _, v := range pod.Spec.Volumes {
				if v.PersistentVolumeClaim != nil {
					pvc, err := cl.CoreV1().PersistentVolumeClaims(pod.Namespace).Get(context.TODO(), v.PersistentVolumeClaim.ClaimName, metav1.GetOptions{})
					if err != nil {
						return false, err
					}
					f, err := checkPersistentVolumeClaim(logger, cl, pvc)
					if err != nil {
						return false, err
					}
					found = found || f
				}
			}
		}
	}

	if pod.Status.Phase != corev1.PodRunning {
		// check events associated with the pod
		f, err := checkEvents(logger, cl, pod)
		if err != nil {
			return false, err
		}
		found = found || f
	}

	return found, nil
}

// check the status of the pod containers
// return the list of containers' name whose status is `waiting`
func diagnoseContainer(logger logr.Logger, cl *kubernetes.Clientset, pod *corev1.Pod) (bool, error) {
	found := false
	for _, s := range pod.Status.ContainerStatuses {
		// if container not in `Running` state
		if s.State.Waiting != nil {
			found = true
			if s.State.Waiting.Message != "" {
				logger.Errorf("👻 container '%s' is waiting with reason '%s': %s", s.Name, s.State.Waiting.Reason, s.State.Waiting.Message)
			} else {
				logger.Errorf("👻 container '%s' is waiting with reason '%s'", s.Name, s.State.Waiting.Reason)
			}
		}
		// also, check the logs
		if (s.Started != nil && *s.Started && !s.Ready) ||
			s.LastTerminationState.Running != nil ||
			s.LastTerminationState.Terminated != nil ||
			s.LastTerminationState.Waiting != nil {
			f, err := checkContainerLogs(logger, cl, pod, s.Name)
			if err != nil {
				return false, err
			}
			found = found || f
		}
		// TODO: check reason and provide a more detailed diagnosis or hint to fix the problem?
		// eg: if reason is `CrashLoopBackOff`, look for errors (`ERROR`/`FATAL`) in the container logs? (but display the n last lines?)
		// eg: if reason is `CreateContainerConfigError`, message should be enough (eg: `secret "cookie" not found`)
	}
	return found, nil
}

func checkContainerLogs(logger logr.Logger, cl *kubernetes.Clientset, pod *corev1.Pod, container string) (bool, error) {
	found := false
	logger.Debugf("👀 checking '%s' container logs...", container)
	logs, err := cl.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: container}).DoRaw(context.TODO())
	if err != nil {
		return false, err
	}
	logger.Debugf("logs: '%s'", string(logs))
	for _, l := range strings.Split(string(logs), "\n") {
		ll := strings.ToLower(l)
		if strings.Contains(ll, "err") ||
			strings.Contains(ll, "fatal") ||
			strings.Contains(ll, "panic") ||
			strings.Contains(ll, "emerg") {
			found = true
			logger.Errorf("🗒  %s", l)
		}
	}
	if !found {
		logger.Infof("🤷 no 'error'/'fatal'/'panic'/'emerg' messages found in the '%s' container logs", container)
	}
	return found, nil
}
