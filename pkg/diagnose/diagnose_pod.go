package diagnose

import (
	"context"
	"fmt"
	"strings"
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
			waiting := checkContainerStatuses(logger, pod)
			found := len(waiting) > 0
			// also, check events associated with the pod
			f, err := checkPodEvents(logger, cfg, pod)
			if err != nil {
				return false, err
			}
			found = found || f
			// also, check logs
			f, err = checkPodLogs(logger, cfg, pod, waiting...)
			if err != nil {
				return false, err
			}
			found = found || f

			return found, nil
		}
	}
	return false, nil
}

// check the status of the pod containers
// return the list of containers' name whose status is `waiting`
func checkContainerStatuses(logger logr.Logger, pod *corev1.Pod) []string {
	waiting := []string{}
	for _, s := range pod.Status.ContainerStatuses {
		// container is not in `Running` state
		if s.State.Waiting != nil {
			logger.Errorf("👻 container '%s' is waiting with reason '%s': %s", s.Name, s.State.Waiting.Reason, s.State.Waiting.Message)
			// TODO: check reason and provide a more detailed diagnosis or hint to fix the problem?
			// if reason is `CrashLoopBackOff`, look for errors (`ERROR`/`FATAL`) in the container logs? (but display the n last lines?)
			// if reason is `CreateContainerConfigError`, message should be enough (eg: `secret "cookie" not found`)
			waiting = append(waiting, s.Name)
		}
	}
	return waiting
}

func checkPodEvents(logger logr.Logger, cfg *rest.Config, pod *corev1.Pod) (bool, error) {
	logger.Infof("👀 checking pod events...")
	cl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return false, err
	}
	events, err := cl.CoreV1().Events(pod.Namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", pod.Name), // TODO: include 'kind'
	})
	if err != nil {
		return false, err
	}
	found := false
	for _, e := range events.Items {
		if e.Type == corev1.EventTypeWarning {
			logger.Infof("⚡️ %s ago: %s", time.Since(e.LastTimestamp.Time).Truncate(time.Second), e.Message)
			found = true
		}
	}
	return found, nil
}

func checkPodLogs(logger logr.Logger, cfg *rest.Config, pod *corev1.Pod, containers ...string) (bool, error) {
	cl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return false, err
	}
	found := false
	for _, container := range containers {
		logger.Infof("👀 checking logs in '%s' container...", container)
		logs, err := cl.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: container}).DoRaw(context.TODO())
		if err != nil {
			return false, err
		}
		logger.Debugf("logs: '%s'", string(logs))
		for _, l := range strings.Split(string(logs), "\n") {
			ll := strings.ToLower(l)
			if strings.Contains(ll, "error") ||
				strings.Contains(ll, "fatal") ||
				strings.Contains(ll, "panic") ||
				strings.Contains(ll, "emerg") {
				found = true
				logger.Errorf("🗒  %s", l)
			}
		}
	}
	if !found {
		logger.Infof("🤷 no relevant message found in the pod logs (but you may want to check yourself)")
	}
	return found, nil
}
