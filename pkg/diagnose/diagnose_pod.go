package diagnose

import (
	"context"
	"fmt"
	"time"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func DiagnoseFromPod(logger logr.Logger, cfg *rest.Config, namespace, name string) (bool, error) {
	pod, err := getPod(cfg, namespace, name)
	if err != nil {
		return false, err
	}
	return checkPod(logger, cfg, pod)
}

func getPod(cfg *rest.Config, namespace, name string) (*corev1.Pod, error) {
	cl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cl.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func checkPod(logger logr.Logger, cfg *rest.Config, pod *corev1.Pod) (bool, error) {
	logger.Infof("👀 checking pod '%s'...", pod.Name)
	return checkPodStatus(logger, cfg, pod)
}

// check the status of the pod status
func checkPodStatus(logger logr.Logger, cfg *rest.Config, pod *corev1.Pod) (bool, error) {
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.ContainersReady && c.Status == corev1.ConditionFalse {
			logger.Errorf("👻 %s", c.Message)
			// also, check the container statuses
			if checkContainerStatuses(logger, pod) {
				return true, nil
			}
			return checkPodEvents(logger, cfg, pod)
		}
	}
	return false, nil
}

// check the status of the pod containers
func checkContainerStatuses(logger logr.Logger, pod *corev1.Pod) bool {
	for _, s := range pod.Status.ContainerStatuses {
		// container is not in `Running` state
		if s.State.Waiting != nil {
			logger.Errorf("👻 container '%s' is waiting with reason '%s': %s", s.Name, s.State.Waiting.Reason, s.State.Waiting.Message)
			// TODO: check reason and provide a more detailed diagnosis or hint to fix the problem?
			// if reason is `CrashLoopBackOff`, look for errors (`ERROR`/`FATAL`) in the container logs? (but display the n last lines?)
			// if reason is `CreateContainerConfigError`, message should be enough (eg: `secret "cookie" not found`)
			return true
		}
	}
	return false
}

func checkPodEvents(logger logr.Logger, cfg *rest.Config, pod *corev1.Pod) (bool, error) {
	logger.Debugf("⚡️ looking for events...")
	cl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return false, err
	}
	events, err := cl.CoreV1().Events(pod.Namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.namespace=%s,involvedObject.name=%s", pod.Namespace, pod.Name),
	})
	if err != nil {
		return false, err
	}
	found := false
	for _, e := range events.Items {
		if e.Type == corev1.EventTypeWarning {
			logger.Infof("⚡️ %s ago: %s", time.Since(e.LastTimestamp.Time).Truncate(time.Second).String(), e.Message)
			found = true
		}
	}
	return found, nil
}
