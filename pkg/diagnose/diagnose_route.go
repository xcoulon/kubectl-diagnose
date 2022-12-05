package diagnose

import (
	"context"

	routev1 "github.com/openshift/api/route/v1"
	routeclient "github.com/openshift/client-go/route/clientset/versioned"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
)

func getRoute(cfg *rest.Config, namespace, name string) (*routev1.Route, error) {
	cl, err := routeclient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cl.RouteV1().Routes(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// checks:
// - the route's target port on pods selected by the service this route points to.
// (If this is a string, it will be looked up as a named port in the target endpoints port list)
func diagnoseRoute(logger logr.Logger, cfg *rest.Config, route *routev1.Route) (bool, error) {
	logger.Infof("👀 checking route '%s' in namespace '%s'...", route.Name, route.Namespace)
	svc, err := getService(cfg, route.Namespace, route.Spec.To.Name)
	if apierrors.IsNotFound(err) {
		logger.Errorf("👻 unable to find service '%s'", route.Spec.To.Name)
		return true, nil
	} else if err != nil {
		return false, err
	}
	// check that the route's `targetPort` matches a `port` on the destination service
	targetPort := route.Spec.Port.TargetPort
	switch targetPort.Type {
	case intstr.Int:
		for _, port := range svc.Spec.Ports {
			if port.Port == targetPort.IntVal {
				return diagnoseService(logger, cfg, svc)
			}
		}
		logger.Errorf("👻 route target port '%d' is not defined in service '%s'", targetPort.IntVal, svc.Name)
		return true, nil
	default:
		for _, port := range svc.Spec.Ports {
			if port.Name == targetPort.StrVal {
				return diagnoseService(logger, cfg, svc)
			}
		}
		logger.Errorf("👻 route target port '%s' is not defined in service '%s'", targetPort.StrVal, svc.Name)
		return true, nil
	}
}
