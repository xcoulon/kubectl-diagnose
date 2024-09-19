package diagnose

import (
	"context"
	"strings"

	"github.com/charmbracelet/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func diagnosePod(ctx context.Context, logger *log.Logger, cfg *rest.Config, namespace, name string) (bool, error) {
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

func checkPod(ctx context.Context, logger *log.Logger, cl *kubernetes.Clientset, pod *corev1.Pod) (bool, error) {
	logger.Debugf("👀 checking pod '%s' in namespace '%s'...", pod.Name, pod.Namespace)
	found := false
	logger.Debugf("👀 checking pod status...")
	//
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.ContainersReady && c.Status == corev1.ConditionTrue {
			logger.Infof("☑️  all containers in pod '%s' are ready", pod.Name)
			// check the latest logs for the container (or sidecar container, eg, a proxy) that matches the port
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
func diagnoseContainer(ctx context.Context, logger *log.Logger, cl *kubernetes.Clientset, pod *corev1.Pod) (bool, error) {
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

		case cs.Started != nil && *cs.Started:
			f, err := checkContainerLogs(ctx, logger, cl, pod, cs.Name)
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

func checkContainerLogs(ctx context.Context, logger *log.Logger, cl *kubernetes.Clientset, pod *corev1.Pod, container string) (bool, error) {
	found := false
	logs, err := cl.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: container}).DoRaw(ctx)
	if err != nil {
		return false, err
	}
	logger.Debugf("logs: '%s'", string(logs))
	for _, l := range strings.Split(string(logs), "\n") {
		// known errors
		if strings.Contains(l, "tls: bad record MAC") {
			logger.Errorf("🗒  %s", l)
			logger.Errorf("👻 wrong TLS secret mounted on the '%s' container?", container)
			return true, nil
		}
		ll := strings.ToLower(l)

		if strings.Contains(ll, "error") ||
			strings.Contains(ll, "failed") ||
			strings.Contains(ll, "fatal") ||
			strings.Contains(ll, "panic") ||
			strings.Contains(ll, "forbidden") ||
			strings.Contains(ll, "cannot") ||
			strings.Contains(ll, "unable") ||
			strings.Contains(ll, "emerg") {
			found = true

		}
	}
	if !found {
		logger.Debugf("🤷 no 'error'/'failed'/'fatal'/'panic'/'emerg' messages found in the '%s' container logs", container)
		return false, nil
	}
	logger.Errorf("🗒 '%s' container logs contains error messages:", container)
	logger.Printf("%s", string(logs))
	return true, nil
}
