package diagnose

func IsRoute(kind string) bool {
	switch kind {
	case "routes", "route", "route.route.openshift.io":
		return true
	default:
		return false
	}
}

func IsService(kind string) bool {
	switch kind {
	case "services", "service", "svc":
		return true
	default:
		return false
	}
}

func IsDeployment(kind string) bool {
	switch kind {
	case "deployments", "deployment", "deploy", "deployment.apps":
		return true
	default:
		return false
	}
}

func IsReplicaSet(kind string) bool {
	switch kind {
	case "replicasets", "replicaset", "rs", "replicaset.apps":
		return true
	default:
		return false
	}
}

func IsPod(kind string) bool {
	switch kind {
	case "pods", "pod", "po":
		return true
	default:
		return false
	}
}
