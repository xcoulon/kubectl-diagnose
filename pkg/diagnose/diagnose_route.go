package diagnose

import (
	"context"

	routev1 "github.com/openshift/api/route/v1"
	routeclient "github.com/openshift/client-go/route/clientset/versioned"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func diagnoseRoute(logger logr.Logger, cfg *rest.Config, namespace, name string) (bool, error) {
	r, err := routeclient.NewForConfigOrDie(cfg).RouteV1().Routes(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	cl := kubernetes.NewForConfigOrDie(cfg)
	return checkRoute(logger, cl, r)
}

// checks:
// - the route's target port on pods selected by the service this route points to.
// (If this is a string, it will be looked up as a named port in the target endpoints port list)
func checkRoute(logger logr.Logger, cl *kubernetes.Clientset, route *routev1.Route) (bool, error) {
	logger.Infof("👀 checking route '%s' in namespace '%s'...", route.Name, route.Namespace)
	svc, err := cl.CoreV1().Services(route.Namespace).Get(context.TODO(), route.Spec.To.Name, metav1.GetOptions{})
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
				return checkService(logger, cl, svc)
			}
		}
		logger.Errorf("👻 route target port '%d' is not defined in service '%s'", targetPort.IntVal, svc.Name)
		return true, nil
	default:
		for _, port := range svc.Spec.Ports {
			if port.Name == targetPort.StrVal {
				return checkService(logger, cl, svc)
			}
		}
		logger.Errorf("👻 route target port '%s' is not defined in service '%s'", targetPort.StrVal, svc.Name)
		return true, nil
	}
}
