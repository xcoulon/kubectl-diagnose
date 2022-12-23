package testsupport_test

import (
	"io/ioutil"
	"net/http"

	"github.com/xcoulon/kubectl-diagnose/testsupport"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("fake api-server endpoints", func() {

	It("should get single pod", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/pods/all-good-785d8bcc5f-g92mn")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&corev1.Pod{}))
	})

	It("should get no pod", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/pods/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should fail to get single pod", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/pods/error")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusInternalServerError))
	})

	It("should list 2 pods", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/pods?labelSelector=app%3Dall-good")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedPodCount(2))
	})

	It("should list no pod", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/pods?labelSelector=app%3Dunknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedPodCount(0))
	})

	It("should get single replicaset", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/deployment-service-account-not-found.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/replicasets/deploy-sa-notfound-59b5d8468f")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&appsv1.ReplicaSet{}))
	})

	It("should get no replicaset", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/replicasets/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should fail to get single replicaset", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/replicasets/error")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusInternalServerError))
	})

	It("should list 2 replicasets", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/replicasets")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedReplicaSetCount(2))
	})

	It("should list 1 replicaset by labelSelector", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/replicasets?labelSelector=app%3Drs-1")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedReplicaSetCount(1))
	})

	It("should list 2 replicasets by labelSelector", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/replicasets?labelSelector=more%3Dcookies")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedReplicaSetCount(2))
	})

	It("should list 2 statefulsets", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/statefulsets")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedStatefulSetCount(2))
	})

	It("should list 1 replicaset by labelSelector", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/statefulsets?labelSelector=app%3Dstatefulset-all-good")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedStatefulSetCount(1))
	})

	It("should get single deployment", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/deployment-service-account-not-found.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/deployments/deploy-sa-notfound")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&appsv1.Deployment{}))
	})

	It("should get no deployment", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/deployments/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should fail to get single deployment", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/deployments/error")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusInternalServerError))
	})

	It("should get single service", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/services/all-good")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&corev1.Service{}))
	})

	It("should get no service", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/services/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should fail to get single service", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/services/error")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusInternalServerError))
	})

	It("should get single route", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/route.openshift.io/v1/namespaces/test/routes/all-good")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&routev1.Route{}))
	})

	It("should get no route", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/route.openshift.io/v1/namespaces/test/routes/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should fail to get single route", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/route.openshift.io/v1/namespaces/test/routes/error")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusInternalServerError))
	})

	It("should get single ingress", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/networking.k8s.io/v1/namespaces/test/ingresses/all-good")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&networkingv1.Ingress{}))
	})

	It("should get no ingress", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/networking.k8s.io/v1/namespaces/test/ingresses/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should fail to get single ingress", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/networking.k8s.io/v1/namespaces/test/ingresses/error")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusInternalServerError))
	})

	It("should get single ingressclass", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/networking.k8s.io/v1/ingressclasses/nginx")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&networkingv1.IngressClass{}))
	})

	It("should get no ingressclass", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/networking.k8s.io/v1/namespaces/test/ingresses/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should fail to get single ingressclass", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/networking.k8s.io/v1/namespaces/test/ingresses/forbidden")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusForbidden))
	})

	It("should list 1 event by fieldSelector", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/events?fieldSelector=involvedObject.name%3Dreadiness-probe-error-6cb7664768-qlmns")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedEventCount(1))
	})

	It("should list no events by fieldSelector", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/events?fieldSelector=involvedObject.name%3Dunknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedEventCount(0))
	})

	It("should retrieve logs", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml", "resources/all-good.logs")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/pods/all-good-785d8bcc5f-g92mn/log?container=default")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveTextBody("some\nlogs"))
	})

	It("should not match any endpoint", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger)
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/nodes/unknown") // unsupported kind of resource: nodes

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should get single pvc", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/persistentvolumeclaims/caddy-config-cache-statefulset-all-good-0")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&corev1.PersistentVolumeClaim{}))
	})

	It("should get no pvc", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/persistentvolumeclaims/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should fail to get single pvc", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/persistentvolumeclaims/error")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusInternalServerError))
	})

	It("should list 1 pvc per label selector", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/persistentvolumeclaims?labelSelector=app%3Dunknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedPersistentVolumeClaimCount(0))
	})

	It("should list 1 pvc per label selector", func() {
		// given
		logger := testsupport.NewLogger()
		s, err := testsupport.NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/persistentvolumeclaims?labelSelector=app%3Dstatefulset-all-good")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedPersistentVolumeClaimCount(1))
	})

})

func HaveBodyOfType(expected runtime.Object) types.GomegaMatcher {
	return And(
		HaveHTTPStatus(200),
		HaveHTTPHeaderWithValue("Content-Type", "application/json"),
		WithTransform(func(resp *http.Response) (runtime.Object, error) {
			deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
			data, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return nil, err
			}
			obj, _, err := deserializer.Decode(data, nil, nil)
			return obj, err
		}, BeAssignableToTypeOf(expected)),
	)
}

func HaveTextBody(expected string) types.GomegaMatcher {
	return And(
		HaveHTTPStatus(200),
		HaveHTTPHeaderWithValue("Content-Type", "text/plain"),
		WithTransform(func(resp *http.Response) (string, error) {
			data, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return "", err
			}
			return string(data), err
		}, Equal(expected)),
	)
}

func HaveReturnedPodCount(expected int) types.GomegaMatcher {
	return And(
		HaveHTTPStatus(200),
		HaveHTTPHeaderWithValue("Content-Type", "application/json"),
		WithTransform(func(resp *http.Response) ([]corev1.Pod, error) {
			deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
			data, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return nil, err
			}
			list := &corev1.PodList{}
			_, _, err = deserializer.Decode(data, nil, list)
			return list.Items, err
		}, HaveLen(expected)),
	)
}

func HaveReturnedPersistentVolumeClaimCount(expected int) types.GomegaMatcher {
	return And(
		HaveHTTPStatus(200),
		HaveHTTPHeaderWithValue("Content-Type", "application/json"),
		WithTransform(func(resp *http.Response) ([]corev1.PersistentVolumeClaim, error) {
			deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
			data, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return nil, err
			}
			list := &corev1.PersistentVolumeClaimList{}
			_, _, err = deserializer.Decode(data, nil, list)
			return list.Items, err
		}, HaveLen(expected)),
	)
}

func HaveReturnedReplicaSetCount(expected int) types.GomegaMatcher {
	return And(
		HaveHTTPStatus(200),
		HaveHTTPHeaderWithValue("Content-Type", "application/json"),
		WithTransform(func(resp *http.Response) ([]appsv1.ReplicaSet, error) {
			deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
			data, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return nil, err
			}
			list := &appsv1.ReplicaSetList{}
			_, _, err = deserializer.Decode(data, nil, list)
			return list.Items, err
		}, HaveLen(expected)),
	)
}

func HaveReturnedStatefulSetCount(expected int) types.GomegaMatcher {
	return And(
		HaveHTTPStatus(200),
		HaveHTTPHeaderWithValue("Content-Type", "application/json"),
		WithTransform(func(resp *http.Response) ([]appsv1.StatefulSet, error) {
			deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
			data, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return nil, err
			}
			list := &appsv1.StatefulSetList{}
			_, _, err = deserializer.Decode(data, nil, list)
			return list.Items, err
		}, HaveLen(expected)),
	)
}

func HaveReturnedEventCount(expected int) types.GomegaMatcher {
	return And(
		HaveHTTPStatus(200),
		HaveHTTPHeaderWithValue("Content-Type", "application/json"),
		WithTransform(func(resp *http.Response) ([]corev1.Event, error) {
			deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
			data, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return nil, err
			}
			list := &corev1.EventList{}
			_, _, err = deserializer.Decode(data, nil, list)
			return list.Items, err
		}, HaveLen(expected)),
	)
}
