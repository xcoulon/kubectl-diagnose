package diagnose

import (
	"fmt"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"

	"k8s.io/client-go/rest"
)

func Diagnose(logger logr.Logger, cfg *rest.Config, kind ResourceKind, namespace, name string) (bool, error) {
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
