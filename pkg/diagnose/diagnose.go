package diagnose

import (
	"fmt"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"

	"k8s.io/client-go/rest"
)

func Diagnose(logger logr.Logger, cfg *rest.Config, kind, namespace, name string) (bool, error) {
	switch {
	case IsRoute(kind):
		r, err := getRoute(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return checkRoute(logger, cfg, r)
	case IsService(kind):
		svc, err := getService(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return checkService(logger, cfg, svc)
	case IsDeployment(kind):
		d, err := getDeployment(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return checkDeployment(logger, cfg, d)
	case IsReplicaSet(kind):
		rs, err := getReplicaSet(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return checkReplicaSet(logger, cfg, rs)
	case IsPod(kind):
		pod, err := getPod(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return checkPod(logger, cfg, pod)
	default:
		return false, fmt.Errorf("🤷 unsupported kind of resource: '%s'", kind)
	}
}
