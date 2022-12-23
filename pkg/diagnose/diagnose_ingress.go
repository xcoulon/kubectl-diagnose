package diagnose

import (
	"context"
	"strconv"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func diagnoseIngress(logger logr.Logger, cfg *rest.Config, namespace, name string) (bool, error) {
	cl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return false, err
	}
	i, err := cl.NetworkingV1().Ingresses(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return checkIngress(logger, cl, i)
}

func checkIngress(logger logr.Logger, cl *kubernetes.Clientset, i *networkingv1.Ingress) (bool, error) {
	logger.Infof("👀 checking ingress '%s' in namespace '%s'...", i.Name, i.Namespace)
	if i.Spec.IngressClassName != nil {
		logger.Infof("`👀 checking ingressclass '%s' at cluster level...`", *i.Spec.IngressClassName)
		// look for ingress classnames (if allowed)
		if _, err := cl.NetworkingV1().IngressClasses().Get(context.TODO(), *i.Spec.IngressClassName, metav1.GetOptions{}); errors.IsNotFound(err) {
			logger.Errorf("👻 unable to find ingressclass '%s'", *i.Spec.IngressClassName)
			return true, nil
		} else if errors.IsForbidden(err) {
			// ingressclasses are cluster-scoped resources and user may not be allowed to get/list such resources
			logger.Infof("🤷 unable to verify ingressclass '%s': %v", *i.Spec.IngressClassName, err)
		} else if err != nil {
			return false, err
		}
	}
	for _, r := range i.Spec.Rules {
		if h := r.HTTP; h != nil {
		paths:
			for _, p := range h.Paths {
				if s := p.Backend.Service; s != nil {
					// look-up service by name
					svc, err := cl.CoreV1().Services(i.Namespace).Get(context.TODO(), s.Name, metav1.GetOptions{})
					if errors.IsNotFound(err) {
						logger.Errorf("👻 unable to find service '%s' associated with host '%s' and path '%s'", s.Name, r.Host, p.Path)
						return true, nil
					} else if err != nil {
						return false, err
					}
					for _, p := range svc.Spec.Ports {
						if s.Port.Number == p.Port || s.Port.Name == p.Name {
							if found, err := checkService(logger, cl, svc); found || err != nil {
								return found, err
							}
							continue paths
						}
					}
					logger.Errorf("👻 port '%s' is not defined in service '%s'", portOrName(s.Port), svc.Name)
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func portOrName(p networkingv1.ServiceBackendPort) string {
	if p.Number != 0 {
		return strconv.Itoa(int(p.Number))
	}
	return p.Name
}
