package diagnose

import (
	"context"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getService(cfg *rest.Config, namespace, name string) (*corev1.Service, error) {
	cl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cl.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func checkService(logger logr.Logger, cfg *rest.Config, svc *corev1.Service) (bool, error) {
	logger.Infof("👀 checking service '%s' in namespace '%s'...", svc.Name, svc.Namespace)
	// Check endpoints
	// endpoints, _ := d.CoreV1().Endpoints(namespace).Get(context.TODO(), name, metav1.GetOptions{})

	// find all pods with the associated label selector in the same namespace
	selector := labels.Set(svc.Spec.Selector).String()
	cl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return false, err
	}
	pods, err := cl.CoreV1().Pods(svc.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return false, err
	}
	// if there is no pod matching the selector
	if len(pods.Items) == 0 {
		logger.Errorf("👻 no pods matching label selector '%s' found in namespace '%s'", selector, svc.Namespace)
		logger.Infof("💡 you may want to:")
		logger.Infof(" - check the 'service.spec.selector' value")
		logger.Infof(" - make sure that the expected pods exists")
		return true, nil
	}
pods:
	for _, pod := range pods.Items {
		for _, sp := range svc.Spec.Ports {
			logger.Debugf("   checking pod '%s'...", pod.Name)
			// check the svc/pod port bindings
		containers:
			for _, c := range pod.Spec.Containers {
				for _, cp := range c.Ports {
					if cp.Name == sp.TargetPort.StrVal || cp.ContainerPort == sp.TargetPort.IntVal {
						logger.Infof("☑️ found matching target port '%s' (%d) in container '%s' of pod '%s'", cp.Name, cp.ContainerPort, c.Name, pod.Name)
						break containers
					}
				}
				logger.Errorf("👻 no container with matching target port '%s' in pod '%s'", sp.TargetPort.String(), pod.Name)
				return true, nil
			}
			p := pod
			if found, err := checkPod(logger, cfg, &p); err != nil {
				return false, err
			} else if found {
				return true, nil
			}
			continue pods
		}
	}

	return false, nil
}
