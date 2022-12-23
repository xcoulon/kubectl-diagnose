package diagnose

import (
	"context"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func diagnosePersistentVolumeClaim(logger logr.Logger, cfg *rest.Config, namespace, name string) (bool, error) {
	cl := kubernetes.NewForConfigOrDie(cfg)
	pvc, err := cl.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return checkPersistentVolumeClaim(logger, cl, pvc)
}

func checkPersistentVolumeClaim(logger logr.Logger, cl *kubernetes.Clientset, pvc *corev1.PersistentVolumeClaim) (bool, error) {
	logger.Infof("👀 checking persistentvolumeclaim '%s' in namespace '%s'...", pvc.Name, pvc.Namespace)
	found := false
	logger.Debugf("👀 checking persistentvolumeclaim status...")
	if pvc.Status.Phase == corev1.ClaimPending {
		logger.Errorf("👻 persistentvolumeclaim '%s' is in '%s' phase", pvc.Name, pvc.Status.Phase)
	}
	// check events associated with the pod
	f, err := checkEvents(logger, cl, pvc)
	if err != nil {
		return false, err
	}
	found = found || f
	return found, nil
}
