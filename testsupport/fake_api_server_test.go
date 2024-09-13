package testsupport_test

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xcoulon/kubectl-diagnose/testsupport"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestFakeAPIServer(t *testing.T) {

	testdata := []struct {
		title     string
		resources []string
		endpoint  string
		expected  func(*http.Response) assert.Comparison
	}{
		{
			title: "should get single pod",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/pods/all-good-785d8bcc5f-g92mn",
			expected: BodyOfType(&corev1.Pod{}),
		},
		{
			title: "should get no pod",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/pods/unknown",
			expected: HTTPStatus(http.StatusNotFound),
		},
		{
			title: "should fail to get single pod",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/pods/error",
			expected: HTTPStatus(http.StatusInternalServerError),
		},
		{
			title: "should list 2 pods",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/pods?labelSelector=app%3Dall-good",
			expected: PodCount(2),
		},
		{
			title: "should list no pod",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/pods?labelSelector=app%3Dunknown",
			expected: PodCount(0),
		},
		{
			title: "should get single replicaset",
			resources: []string{
				"resources/deployment-service-account-not-found.yaml",
			},
			endpoint: "/apis/apps/v1/namespaces/test/replicasets/deploy-sa-notfound-59b5d8468f",
			expected: BodyOfType(&appsv1.ReplicaSet{}),
		},
		{
			title: "should get no replicaset",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/apps/v1/namespaces/test/replicasets/unknown",
			expected: HTTPStatus(http.StatusNotFound),
		},
		{
			title: "should fail to get single replicaset",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/apps/v1/namespaces/test/replicasets/error",
			expected: HTTPStatus(http.StatusInternalServerError),
		},
		{
			title: "should list 2 replicasets",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/apps/v1/namespaces/test/replicasets",
			expected: ReplicaSetCount(2),
		},
		{
			title: "should list 1 replicaset by labelSelector",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/apps/v1/namespaces/test/replicasets?labelSelector=app%3Drs-1",
			expected: ReplicaSetCount(1),
		},
		{
			title: "should list 2 replicasets by labelSelector",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/apps/v1/namespaces/test/replicasets?labelSelector=more%3Dcookies",
			expected: ReplicaSetCount(2),
		},
		{
			title: "should list 2 statefulsets",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/apps/v1/namespaces/test/statefulsets",
			expected: StatefulSetCount(2),
		},
		{
			title: "should list 1 statefulsets by labelSelector",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/apps/v1/namespaces/test/statefulsets?labelSelector=app%3Dstatefulset-all-good",
			expected: StatefulSetCount(1),
		},
		{
			title: "should get single deployment",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/apps/v1/namespaces/test/deployments/deploy-all-good-en",
			expected: BodyOfType(&appsv1.Deployment{}),
		},
		{
			title: "should get no deployment",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/apps/v1/namespaces/test/deployments/unknown",
			expected: HTTPStatus(http.StatusNotFound),
		},
		{
			title: "should fail to get single deployment",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/apps/v1/namespaces/test/deployments/error",
			expected: HTTPStatus(http.StatusInternalServerError),
		},
		{
			title: "should list 1 deployment by labelSelector",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/apps/v1/namespaces/test/deployments?labelSelector=app%3Ddeploy-all-good",
			expected: StatefulSetCount(1),
		},
		{
			title: "should get single service",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/services/all-good",
			expected: BodyOfType(&corev1.Service{}),
		},
		{
			title: "should get no service",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/services/unknown",
			expected: HTTPStatus(http.StatusNotFound),
		},
		{
			title: "should fail to get single service",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/services/error",
			expected: HTTPStatus(http.StatusInternalServerError),
		},
		{
			title: "should get single route",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/route.openshift.io/v1/namespaces/test/routes/all-good",
			expected: BodyOfType(&routev1.Route{}),
		},
		{
			title: "should get no route",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/route.openshift.io/v1/namespaces/test/routes/unknown",
			expected: HTTPStatus(http.StatusNotFound),
		},
		{
			title: "should fail to get single route",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/route.openshift.io/v1/namespaces/test/routes/error",
			expected: HTTPStatus(http.StatusInternalServerError),
		},
		{
			title: "should get single ingress",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/networking.k8s.io/v1/namespaces/test/ingresses/all-good",
			expected: BodyOfType(&networkingv1.Ingress{}),
		},
		{
			title: "should get no ingress",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/networking.k8s.io/v1/namespaces/test/ingresses/unknown",
			expected: HTTPStatus(http.StatusNotFound),
		},
		{
			title: "should fail to get single ingress",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/networking.k8s.io/v1/namespaces/test/ingresses/error",
			expected: HTTPStatus(http.StatusInternalServerError),
		},
		{
			title: "should get single ingressclass",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/networking.k8s.io/v1/ingressclasses/nginx",
			expected: BodyOfType(&networkingv1.IngressClass{}),
		},
		{
			title: "should get no ingressclass",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/networking.k8s.io/v1/namespaces/test/ingresses/unknown",
			expected: HTTPStatus(http.StatusNotFound),
		},
		{
			title: "should fail to get single ingressclass",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/apis/networking.k8s.io/v1/namespaces/test/ingresses/forbidden",
			expected: HTTPStatus(http.StatusForbidden),
		},
		{
			title: "should list 1 event by fieldSelector",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/events?fieldSelector=type=Warning,involvedObject.uid%3Dc188fba1-0304-4686-a67c-17db71548c6f,involvedObject.resourceVersion%3D277004",
			expected: EventCount(1),
		},
		{
			title: "should list no events by fieldSelector",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/events?fieldSelector=involvedObject.name%3Dunknown",
			expected: EventCount(0),
		},
		{
			title: "should retrieve logs",
			resources: []string{
				"resources/fake-api-server.yaml",
				"resources/fake-api-server.logs",
			},
			endpoint: "/api/v1/namespaces/test/pods/all-good-785d8bcc5f-g92mn/log?container=default",
			expected: TextBody("some\nlogs"),
		},
		{
			title: "should fail to retrieve logs when container creating",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/pods/container-creating/log?container=default",
			expected: HTTPStatus(http.StatusInternalServerError), // plus `HTTPBody("container 'default' in pod 'container-creating' is waiting to start: ContainerCreating")`
		},
		{
			title: "should not match any endpoint",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/nodes/unknown", // unsupported kind of resource: nodes
			expected: HTTPStatus(http.StatusNotFound),
		},
		{
			title: "should get single pvc",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/persistentvolumeclaims/caddy-config-cache-statefulset-all-good-0",
			expected: BodyOfType(&corev1.PersistentVolumeClaim{}),
		},
		{
			title: "should get no pvc",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/persistentvolumeclaims/unknown",
			expected: HTTPStatus(http.StatusNotFound),
		},
		{
			title: "should fail to get single pvc",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/persistentvolumeclaims/error",
			expected: HTTPStatus(http.StatusInternalServerError),
		},
		{
			title: "should list 1 pvc per label selector",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/persistentvolumeclaims?labelSelector=app%3Dunknown",
			expected: PersistentVolumeClaimCount(0),
		},
		{
			title: "should list 1 pvc per label selector",
			resources: []string{
				"resources/fake-api-server.yaml",
			},
			endpoint: "/api/v1/namespaces/test/persistentvolumeclaims?labelSelector=app%3Dstatefulset-all-good",
			expected: PersistentVolumeClaimCount(1),
		},
	}

	for _, tc := range testdata {
		t.Run(tc.title, func(t *testing.T) {
			// given
			buf := &strings.Builder{}
			logger := log.NewWithOptions(buf, log.Options{
				TimeFormat: time.Kitchen,
				Level:      log.DebugLevel,
			})

			s, err := testsupport.NewFakeAPIServer(logger, tc.resources...)
			require.NoError(t, err)
			defer s.Close()

			// when
			req, err := http.NewRequestWithContext(context.TODO(), "GET", s.URL+tc.endpoint, nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)

			// then
			require.NoError(t, err)
			assert.Condition(t, tc.expected(resp))
		})
	}
}

func HTTPStatus(expected int) func(*http.Response) assert.Comparison {
	return func(resp *http.Response) assert.Comparison {
		return func() (success bool) {
			defer resp.Body.Close()
			return resp.StatusCode == expected
		}
	}
}

func BodyOfType(expected runtime.Object) func(*http.Response) assert.Comparison {
	return func(resp *http.Response) assert.Comparison {
		return func() (success bool) {
			if resp.StatusCode != http.StatusOK {
				return false
			}
			if resp.Header.Get("Content-Type") != "application/json" {
				return false
			}
			deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
			data, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return false
			}
			obj, _, err := deserializer.Decode(data, nil, nil)
			if err != nil {
				return false
			}
			return reflect.TypeOf(expected).AssignableTo(reflect.TypeOf(obj))
		}
	}
}

func TextBody(expected string) func(*http.Response) assert.Comparison {
	return func(resp *http.Response) assert.Comparison {
		return func() (success bool) {
			if resp.StatusCode != http.StatusOK {
				return false
			}
			if resp.Header.Get("Content-Type") != "text/plain" {
				return false
			}
			data, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return false
			}
			return string(data) == expected
		}
	}
}

func PodCount(expected int) func(*http.Response) assert.Comparison {
	return func(resp *http.Response) assert.Comparison {
		return func() (success bool) {
			if resp.StatusCode != http.StatusOK {
				return false
			}
			if resp.Header.Get("Content-Type") != "application/json" {
				return false
			}
			deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
			data, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return false
			}
			list := &corev1.PodList{}
			_, _, err = deserializer.Decode(data, nil, list)
			if err != nil {
				return false
			}
			return len(list.Items) == expected
		}
	}
}

func PersistentVolumeClaimCount(expected int) func(*http.Response) assert.Comparison {
	return func(resp *http.Response) assert.Comparison {
		return func() (success bool) {
			if resp.StatusCode != http.StatusOK {
				return false
			}
			if resp.Header.Get("Content-Type") != "application/json" {
				return false
			}
			deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
			data, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return false
			}
			list := &corev1.PersistentVolumeClaimList{}
			_, _, err = deserializer.Decode(data, nil, list)
			if err != nil {
				return false
			}
			return len(list.Items) == expected
		}
	}
}

func ReplicaSetCount(expected int) func(*http.Response) assert.Comparison {
	return func(resp *http.Response) assert.Comparison {
		return func() (success bool) {
			if resp.StatusCode != http.StatusOK {
				return false
			}
			if resp.Header.Get("Content-Type") != "application/json" {
				return false
			}
			deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
			data, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return false
			}
			list := &appsv1.ReplicaSetList{}
			_, _, err = deserializer.Decode(data, nil, list)
			if err != nil {
				return false
			}
			return len(list.Items) == expected
		}
	}
}

func StatefulSetCount(expected int) func(*http.Response) assert.Comparison {
	return func(resp *http.Response) assert.Comparison {
		return func() (success bool) {
			if resp.StatusCode != http.StatusOK {
				return false
			}
			if resp.Header.Get("Content-Type") != "application/json" {
				return false
			}
			deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
			data, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return false
			}
			list := &appsv1.StatefulSetList{}
			_, _, err = deserializer.Decode(data, nil, list)
			if err != nil {
				return false
			}
			return len(list.Items) == expected
		}
	}
}

func EventCount(expected int) func(*http.Response) assert.Comparison {
	return func(resp *http.Response) assert.Comparison {
		return func() (success bool) {
			if resp.StatusCode != http.StatusOK {
				return false
			}
			if resp.Header.Get("Content-Type") != "application/json" {
				return false
			}
			deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
			data, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return false
			}
			list := &corev1.EventList{}
			_, _, err = deserializer.Decode(data, nil, list)
			if err != nil {
				return false
			}
			return len(list.Items) == expected
		}
	}
}
