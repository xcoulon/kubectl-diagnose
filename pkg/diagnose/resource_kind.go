package diagnose

type ResourceKind int

const (
	Route ResourceKind = iota
	Service
	Deployment
	ReplicaSet
	Pod
	StatefulSet
	PersistentVolumeClaim
	StorageClass
	Unkwown
)

func Kind(kind string) ResourceKind {
	switch kind {
	case "routes", "route", "route.route.openshift.io":
		return Route
	case "services", "service", "svc":
		return Service
	case "deployments", "deployment", "deploy", "deployment.apps":
		return Deployment
	case "replicasets", "replicaset", "rs", "replicaset.apps":
		return ReplicaSet
	case "pods", "pod", "po":
		return Pod
	case "statefulsets", "statefulset", "sts", "statefulset.apps":
		return StatefulSet
	case "persistentvolumeclaims", "persistentvolumeclaim", "pvc":
		return PersistentVolumeClaim
	case "storageclasses", "storageclass", "sc", "storageclass.storage.k8s.io":
		return StorageClass
	default:
		return Unkwown
	}
}
