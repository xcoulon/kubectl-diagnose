package test

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
		s, err := NewFakeAPIServer(logger, "resources/all-good.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/default/pods/all-good-785d8bcc5f-g92mn")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&corev1.Pod{}))
	})

	It("should get no pod", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/all-good.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/default/pods/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should list 2 pods", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/all-good.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/default/pods?labelSelector=app%3Dall-good")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedPodCount(2))
	})

	It("should get single replicaset", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/replicaset-service-account-not-found.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/default/replicasets/sa-notfound")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&appsv1.ReplicaSet{}))
	})

	It("should get no replicaset", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/all-good.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/apps/v1/namespaces/default/replicasets/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should get single service", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/service-invalid-target-port.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/default/services/service-invalid-target-port")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&corev1.Service{}))
	})

	It("should get no service", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/all-good.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/default/services/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should get single route", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/all-good.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/route.openshift.io/v1/namespaces/default/routes/all-good")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveBodyOfType(&routev1.Route{}))
	})

	It("should get no route", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/all-good.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/apis/route.openshift.io/v1/namespaces/default/routes/unknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
	})

	It("should get 1 event", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/pod-readiness-probe-error.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/default/events?fieldSelector=involvedObject.name%3Dreadiness-probe-error")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedEventCount(1))
	})

	It("should get no events", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/pod-readiness-probe-error.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/default/events?fieldSelector=involvedObject.name%3Dunknown")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveReturnedEventCount(0))
	})

	It("should retrieve logs", func() {
		// given
		logger := logr.New(io.Discard)
		s, err := NewFakeAPIServer(logger, "resources/all-good.yaml", "resources/all-good.logs")
		Expect(err).NotTo(HaveOccurred())
		defer s.Close()

		// when
		resp, err := http.DefaultClient.Get(s.URL + "/api/v1/namespaces/default/pods/all-good-785d8bcc5f-g92mn/log?container=default")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveTextBody("some\nlogs"))
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
