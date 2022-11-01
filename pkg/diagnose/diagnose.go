package diagnose

import (
	"fmt"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	"k8s.io/client-go/rest"
)

func Diagnose(logger logr.Logger, cfg *rest.Config, kind, namespace, name string) (bool, error) {
	switch kind {
	case "route":
		return DiagnoseFromRoute(logger, cfg, namespace, name)
	case "service", "svc":
		return DiagnoseFromService(logger, cfg, namespace, name)
	case "replicaset", "rs":
		return DiagnoseFromReplicaSet(logger, cfg, namespace, name)
	case "pod":
		return DiagnoseFromPod(logger, cfg, namespace, name)
	default:
		return false, fmt.Errorf("🤷 unsupported kind of resource: '%s'", kind)
	}
}
