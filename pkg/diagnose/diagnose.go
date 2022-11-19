package diagnose

import (
	"fmt"
	"strings"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"

	"k8s.io/client-go/rest"
)

// TODO: support aliases and fully qualified names (eg: `po` and `route.route.openshift.io`)
const Pod = "pod"
const Service = "service"
const ServiceShortName = "svc"
const ReplicaSet = "replicaset"
const ReplicaSetShortName = "rs"
const Deployment = "deployment"
const Route = "route"

func Diagnose(logger logr.Logger, cfg *rest.Config, kind, namespace, name string) (bool, error) {
	switch strings.ToLower(kind) {
	case Route:
		r, err := getRoute(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return checkRoute(logger, cfg, r)
	case Service, ServiceShortName:
		svc, err := getService(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return checkService(logger, cfg, svc)
	case Deployment:
		d, err := getDeployment(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return checkDeployment(logger, cfg, d)
	case ReplicaSet, ReplicaSetShortName:
		rs, err := getReplicaSet(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return checkReplicaSet(logger, cfg, rs)
	case Pod:
		pod, err := getPod(cfg, namespace, name)
		if err != nil {
			return false, err
		}
		return checkPod(logger, cfg, pod)
	default:
		return false, fmt.Errorf("🤷 unsupported kind of resource: '%s'", kind)
	}
}
