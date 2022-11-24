package testsupport

import (
	"io"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("fake api-server endpoints", func() {

	It("should get single pod", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
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
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/pods/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should list 2 pods", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
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
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
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
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/deployment-service-account-not-found.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/replicasets/sa-notfound-59b5d8468f")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&appsv1.ReplicaSet{}))
	})

	It("should get no replicaset", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/replicasets/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should list 2 replicasets", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
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
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
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
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/replicasets?labelSelector=more%3Dcookies")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedReplicaSetCount(2))
	})

	It("should get single deployment", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/deployment-service-account-not-found.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/test/deployments/sa-notfound")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&appsv1.Deployment{}))
	})

	It("should get single service", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
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
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/services/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should get single route", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
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
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/fake-api-server.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/route.openshift.io/v1/namespaces/test/routes/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should list 1 event by fieldSelector", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/pod-readiness-probe-error.yaml")
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
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/pod-readiness-probe-error.yaml")
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
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/fake-api-server.yaml", "resources/fake-api-server.logs")
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
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger)
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/test/nodes/unknown") // unsupported kind of resource: nodes

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
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
