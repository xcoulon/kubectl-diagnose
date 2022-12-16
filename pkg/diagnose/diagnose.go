package diagnose

import (
	"fmt"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"

	"k8s.io/client-go/rest"
)

func Diagnose(logger logr.Logger, cfg *rest.Config, kind, namespace, name string) (bool, error) {
	switch Kind(kind) {
	case Route:
		r, err := getRoute(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return diagnoseRoute(logger, cfg, r)
	case Service:
		svc, err := getService(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return diagnoseService(logger, cfg, svc)
	case Deployment:
		d, err := getDeployment(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return diagnoseDeployment(logger, cfg, d)
	case ReplicaSet:
		rs, err := getReplicaSet(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return diagnoseReplicaSet(logger, cfg, rs)
	case Pod:
		pod, err := getPod(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return diagnosePod(logger, cfg, pod)
	case StatefulSet:
		sts, err := getStatefulSet(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return diagnoseStatefulSet(logger, cfg, sts)
	case PersistentVolumeClaim:
		pvc, err := getPersistentVolumeClaim(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return diagnosePersistentVolumeClaim(logger, cfg, pvc)
	default:
		return false, fmt.Errorf("🤷 unsupported kind of resource: '%s'", kind)
	}
}
