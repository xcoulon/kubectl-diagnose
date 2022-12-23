package diagnose

import (
	"context"
	"fmt"
	"time"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func checkEvents(logger logr.Logger, cl *kubernetes.Clientset, obj runtimeclient.Object) (bool, error) {
	logger.Debugf("👀 checking events...")
	events, err := cl.CoreV1().Events(obj.GetNamespace()).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", obj.GetName()), // TODO: include 'Kind' or just use object.UID
	})
	if err != nil {
		return false, err
	}
	found := false
	for _, e := range events.Items {
		if e.Type == corev1.EventTypeWarning {
			logger.Errorf("⚡️ %s ago: %s: %s", time.Since(getTime(e)).Truncate(time.Second), e.Reason, e.Message)
			found = true
		}
	}
	return found, nil
}

func getTime(e corev1.Event) time.Time {
	switch {
	case !e.LastTimestamp.IsZero():
		return e.LastTimestamp.Time
	case !e.FirstTimestamp.IsZero():
		return e.FirstTimestamp.Time
	default:
		return e.EventTime.Time
	}
}
