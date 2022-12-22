package diagnose

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourceKind string

const (
	Route                 = ResourceKind("Route")
	Ingress               = ResourceKind("Ingress")
	IngressClass          = ResourceKind("IngressClass")
	Service               = ResourceKind("Service")
	Deployment            = ResourceKind("Deployment")
	ReplicaSet            = ResourceKind("ReplicaSet")
	Pod                   = ResourceKind("Pod")
	StatefulSet           = ResourceKind("StatefulSet")
	PersistentVolumeClaim = ResourceKind("PersistentVolumeClaim")
	StorageClass          = ResourceKind("StorageClass")
	Unknown               = ResourceKind("Unknown")
)

func NewResourceKind(kind string) ResourceKind {
	switch strings.ToLower(kind) {
	case "routes", "route", "route.route.openshift.io":
		return Route
	case "ingresses", "ingress", "ing", "ingress.networking.k8s.io":
		return Ingress
	case "ingressclasses", "ingressclass", "ingressclass.networking.k8s.io":
		return IngressClass
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
		return Unknown
	}
}

func (k ResourceKind) Matches(kind schema.ObjectKind) bool {
	return k == ResourceKind(kind.GroupVersionKind().Kind)
}

func (k ResourceKind) String() string {
	return strings.ToLower(string(k))
}
