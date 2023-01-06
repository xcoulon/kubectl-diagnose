package diagnose

import (
	"fmt"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
)

func Diagnose(logger logr.Logger, cfg *rest.Config, kind ResourceKind, namespace, name string) (bool, error) {
	found, err := diagnose(logger, cfg, kind, namespace, name)
	switch {
	case apierrors.IsNotFound(err):
		// if resource was not found, just print the error but
		// no need to print of the cmd usage
		logger.Errorf(err.Error())
	case !found:
		logger.Infof(notFoundMsg)
	}
	return found, err
}

func diagnose(logger logr.Logger, cfg *rest.Config, kind ResourceKind, namespace, name string) (bool, error) {
	switch kind {
	case Route:
		return diagnoseRoute(logger, cfg, namespace, name)
	case Ingress:
		return diagnoseIngress(logger, cfg, namespace, name)
	case Service:
		return diagnoseService(logger, cfg, namespace, name)
	case Deployment:
		return diagnoseDeployment(logger, cfg, namespace, name)
	case ReplicaSet:
		return diagnoseReplicaSet(logger, cfg, namespace, name)
	case Pod:
		return diagnosePod(logger, cfg, namespace, name)
	case StatefulSet:
		return diagnoseStatefulSet(logger, cfg, namespace, name)
	case PersistentVolumeClaim:
		return diagnosePersistentVolumeClaim(logger, cfg, namespace, name)
	default:
		return false, fmt.Errorf("🤷 unsupported kind of resource: '%s'", kind)
	}
}

const notFoundMsg = `🤷 couldn't find the culprit
possible causes:
- invalid configuration of a sidecar container or a proxy within the pod
- trying to connect to a container which is listening to '127.0.0.1' instead of '0.0.0.0'
- something else?
`
