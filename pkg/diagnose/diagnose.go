package diagnose

import (
	"context"
	"fmt"

	charmlog "github.com/charmbracelet/log"
	"k8s.io/client-go/rest"
)

func Diagnose(ctx context.Context, logger *charmlog.Logger, cfg *rest.Config, kind ResourceKind, namespace, name string) (bool, error) {
	found, err := diagnose(ctx, logger, cfg, kind, namespace, name)
	if err == nil && !found {
		logger.Infof(NotFoundMsg)
	}
	return found, err
}

func diagnose(ctx context.Context, logger *charmlog.Logger, cfg *rest.Config, kind ResourceKind, namespace, name string) (bool, error) {
	switch kind {
	case Route:
		return diagnoseRoute(ctx, logger, cfg, namespace, name)
	case Ingress:
		return diagnoseIngress(ctx, logger, cfg, namespace, name)
	case Service:
		return diagnoseService(ctx, logger, cfg, namespace, name)
	case Deployment:
		return diagnoseDeployment(ctx, logger, cfg, namespace, name)
	case ReplicaSet:
		return diagnoseReplicaSet(ctx, logger, cfg, namespace, name)
	case Pod:
		return diagnosePod(ctx, logger, cfg, namespace, name)
	case StatefulSet:
		return diagnoseStatefulSet(ctx, logger, cfg, namespace, name)
	case PersistentVolumeClaim:
		return diagnosePersistentVolumeClaim(ctx, logger, cfg, namespace, name)
	default:
		return false, fmt.Errorf("🤷 unsupported kind of resource: '%s'", kind)
	}
}

const NotFoundMsg = `🤷 couldn't find the culprit
💡 possible causes:
   - missing TLS annotations on a route or a service?
   - invalid configuration of a container within the pod?
   - trying to connect to a container listening to '127.0.0.1' instead of '0.0.0.0'?
   - redirecting to an invalid callback URL after logging in on a third-party SSO?
   - something else?`
