package diagnose

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const Now = "now"

func checkEvents(ctx context.Context, logger *log.Logger, cl *kubernetes.Clientset, obj runtimeclient.Object) (bool, error) {
	logger.Debugf("👀 checking events...")
	events, err := cl.CoreV1().Events(obj.GetNamespace()).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("type=Warning,involvedObject.uid=%s", obj.GetUID()),
	})
	if err != nil {
		return false, err
	}
	found := false
	sort.Slice(events.Items, func(i, j int) bool {
		return getTime(events.Items[i]).Before(getTime(events.Items[j]))
	})
	for _, e := range events.Items {
		logger.Errorf("⚡️ %s ago: %s: %s", now(ctx).Sub(getTime(e)).Truncate(time.Second), e.Reason, e.Message)
		found = true
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
