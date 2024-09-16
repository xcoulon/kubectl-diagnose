package diagnose

import (
	"context"

	"github.com/charmbracelet/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func diagnoseService(ctx context.Context, logger *log.Logger, cfg *rest.Config, namespace, name string) (bool, error) {
	cl := kubernetes.NewForConfigOrDie(cfg)
	svc, err := cl.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return checkService(ctx, logger, cl, svc)
}

func checkService(ctx context.Context, logger *log.Logger, cl *kubernetes.Clientset, svc *corev1.Service) (bool, error) {
	logger.Infof("👀 checking service '%s' in namespace '%s'...", svc.Name, svc.Namespace)

	pods, err := findPods(ctx, cl, svc.Namespace, svc.Spec.Selector)
	if err != nil {
		return false, err
	}

	// if there is no pod matching the selector
	if len(pods) == 0 {
		sel := labels.Set(svc.Spec.Selector).AsSelector()
		// attempt to find the Deployment which was supposed to create the Pods (if there is one)
		deploys, err := cl.AppsV1().Deployments(svc.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, deploy := range deploys.Items {
			sel := labels.Set(svc.Spec.Selector).AsSelector()
			if sel.Matches(labels.Set(deploy.Spec.Selector.MatchLabels)) {
				obj := deploy
				found, err := checkDeployment(ctx, logger, cl, &obj)
				if err != nil {
					return false, err
				}
				if found {
					return true, err
				}
			}
		}
		// attempt to find the StatefulSet which was supposed to create the Pods (if there is one)
		stss, err := cl.AppsV1().StatefulSets(svc.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, rs := range stss.Items {
			if sel.Matches(labels.Set(rs.Spec.Selector.MatchLabels)) {
				obj := rs
				found, err := checkStatefulSet(ctx, logger, cl, &obj)
				if err != nil {
					return false, err
				}
				if found {
					return true, err
				}
			}
		}
		logger.Errorf("👻 no pods matching label selector '%s' found in namespace '%s'", sel.String(), svc.Namespace)
		logger.Infof("💡 you may want to verify that the pods exist and their labels match '%s'", sel.String())
		return true, nil
	}
pods:
	for _, pod := range pods {
		logger.Debugf("👀 checking pod '%s'...", pod.Name)
		for _, sp := range svc.Spec.Ports {
			// check the svc/pod port bindings
			found := false
			var targetContainer string
		containers:
			for _, c := range pod.Spec.Containers {
				for _, cp := range c.Ports {
					if cp.Name == sp.TargetPort.StrVal || cp.ContainerPort == sp.TargetPort.IntVal {
						logger.Debugf("☑️ found matching target port '%s' (%d) in container '%s' of pod '%s'", cp.Name, cp.ContainerPort, c.Name, pod.Name)
						found = true
						targetContainer = c.Name
						break containers
					}
				}
			}
			if !found {
				logger.Errorf("👻 no container with port '%s' in pod '%s'", sp.TargetPort.String(), pod.Name)
				return true, nil
			}
			p := pod
			if found, err := checkPod(ctx, logger, cl, &p); err != nil {
				return false, err
			} else if found {
				return true, nil
			}
			if found, err := checkContainerLogs(ctx, logger, cl, &p, targetContainer); err != nil {
				return false, err
			} else if found {
				return true, nil
			}
			continue pods
		}
	}

	return false, nil
}
