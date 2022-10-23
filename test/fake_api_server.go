package test

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed resources/*
var Resources embed.FS

func init() {
	corev1.AddToScheme(scheme.Scheme)  // Kubernetes Pods and Services
	appsv1.AddToScheme(scheme.Scheme)  // Kubernetes ReplicaSets
	routev1.AddToScheme(scheme.Scheme) // OpenShift Routes
}

func NewFakeAPIServer(logger logr.Logger, filename string) (*httptest.Server, []runtimeclient.Object, error) {
	objs, err := parseObjects(filename)
	if err != nil {
		return nil, nil, err
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		logger.Debugf("processing %s %v\n", req.Method, req.URL)
		switch req.Method {
		case "GET":
			obj, err := getResource(logger, objs, req.URL)
			if err != nil {
				logger.Errorf(err.Error())
				w.WriteHeader(http.StatusNotFound)
				return
			}
			output, _ := json.Marshal(obj)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(output) // nolint: errcheck
		default:
			logger.Errorf("unexpected request: %s %s\n", req.Method, req.URL)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	return httptest.NewServer(handler), objs, nil
}

var routeRegexp = regexp.MustCompile(`^/apis/route\.openshift\.io/v1/namespaces/(?P<namespace>[a-z0-9\.-]+)/routes/(?P<name>[a-z0-9\.-]+)$`)
var serviceRegexp = regexp.MustCompile(`^/api/v1/namespaces/(?P<namespace>[a-z0-9\.-]+)/services/(?P<name>[a-z0-9\.-]+)$`)
var replicasetRegexp = regexp.MustCompile(`^/apis/apps/v1/namespaces/(?P<namespace>[a-z0-9\.-]+)/replicasets/(?P<name>[a-z0-9\.-]+)$`)
var podRegexp = regexp.MustCompile(`^/api/v1/namespaces/(?P<namespace>[a-z0-9\.-]+)/pods/(?P<name>[a-z0-9\.-]+)$`)
var podsRegex = regexp.MustCompile(`^/api/v1/namespaces/(?P<namespace>[a-z0-9\.-]+)/pods\?labelSelector=(?P<labelSelector>[a-zA-Z0-9\.%-]+)$`)
var eventsRegex = regexp.MustCompile(`^/api/v1/namespaces/(?P<namespace>[a-z0-9\.-]+)/events\?fieldSelector=(?P<fieldSelector>[a-zA-Z0-9\.%-]+)$`)

func getResource(logger logr.Logger, objs []runtimeclient.Object, u *url.URL) (runtime.Object, error) {
	// get a single resource by kind/namespace/name
	for kind, re := range map[string]*regexp.Regexp{
		"Route":      routeRegexp,
		"Pod":        podRegexp,
		"Service":    serviceRegexp,
		"ReplicaSet": replicasetRegexp,
	} {
		if re.MatchString(u.Path) {
			groups := re.FindStringSubmatch(u.Path)
			namespace := groups[re.SubexpIndex("namespace")]
			name := groups[re.SubexpIndex("name")]
			return getObject(logger, objs, kind, namespace, name)
		}
	}
	// list pods by namespace and label selector
	if pods, err := listPods(logger, objs, u); err != nil {
		return nil, err
	} else if pods != nil {
		return pods, nil
	}
	if events, err := listEvents(logger, objs, u); err != nil {
		return nil, err
	} else if events != nil {
		return events, nil
	}
	return nil, fmt.Errorf("no match for '%s' (invalid path?)", u.String())

}

func getObject(logger logr.Logger, objs []runtimeclient.Object, kind, namespace, name string) (runtimeclient.Object, error) {
	// lookup the object
	logger.Debugf("looking up %s %s/%s", kind, namespace, name)
	for _, obj := range objs {
		if obj.GetObjectKind().GroupVersionKind().Kind == kind &&
			obj.GetNamespace() == namespace &&
			obj.GetName() == name {
			return obj, nil
		}
	}
	return nil, fmt.Errorf("no match for %s %s/%s (missing resource?)", kind, namespace, name)
}

func listPods(logger logr.Logger, objs []runtimeclient.Object, u *url.URL) (*corev1.PodList, error) {
	// lookup the object
	if !podsRegex.MatchString(u.String()) {
		return nil, nil
	}
	groups := podsRegex.FindStringSubmatch(u.String())
	namespace := groups[podsRegex.SubexpIndex("namespace")]
	selector := u.Query()["labelSelector"][0]
	logger.Debugf("listing pods in %s with labels %s", namespace, selector)
	s, err := labels.Parse(selector)
	if err != nil {
		return nil, err
	}
	pods := &corev1.PodList{
		Items: []corev1.Pod{},
	}
	for _, obj := range objs {
		if obj, ok := obj.(*corev1.Pod); ok &&
			obj.GetNamespace() == namespace &&
			s.Matches(labels.Set(obj.GetLabels())) {
			pods.Items = append(pods.Items, *obj)
		}
	}
	return pods, nil
}

func listEvents(logger logr.Logger, objs []runtimeclient.Object, u *url.URL) (*corev1.EventList, error) {
	// lookup the object
	if !eventsRegex.MatchString(u.String()) {
		return nil, nil
	}
	groups := eventsRegex.FindStringSubmatch(u.String())
	namespace := groups[eventsRegex.SubexpIndex("namespace")]
	selector := u.Query()["fieldSelector"][0]
	logger.Debugf("listing pods in %s with labels %s", namespace, selector)
	s, err := fields.ParseSelector(selector)
	if err != nil {
		return nil, err
	}
	events := &corev1.EventList{
		Items: []corev1.Event{},
	}
	for _, obj := range objs {
		if obj, ok := obj.(*corev1.Event); ok &&
			obj.GetNamespace() == namespace &&
			s.Matches(fields.Set(map[string]string{
				"involvedObject.namespace": obj.InvolvedObject.Namespace,
				"involvedObject.name":      obj.InvolvedObject.Name,
			})) {
			events.Items = append(events.Items, *obj)
		}
	}
	return events, nil
}

// see https://github.com/go-yaml/yaml/pull/301#issuecomment-792871300
func parseObjects(filename string) ([]runtimeclient.Object, error) {
	content, err := Resources.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
	objs := []runtimeclient.Object{}
	// decode 1 yaml value at a time, marshal it again and deserialize to `runtime.Object`
	for {
		var value interface{}
		if err := decoder.Decode(&value); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		data, err := yaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		obj, _, err := deserializer.Decode(data, nil, nil)
		if err != nil {
			return nil, err
		}
		if obj, ok := obj.(runtimeclient.Object); ok {
			objs = append(objs, obj)
		}
	}
	return objs, nil
}

var _ = Describe("parse resources", func() {

	It("should parse resources", func() {
		// when
		objects, err := parseObjects("resources/service-invalid-target-port.yaml")
		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(objects).To(HaveLen(2))
	})
})

var _ = DescribeTable("single resource request regexps",
	func(re *regexp.Regexp, path, namespace, name string) {
		match := re.MatchString(path)
		// then
		Expect(match).To(BeTrue())
		groups := re.FindStringSubmatch(path)
		Expect(namespace).To(Equal(groups[re.SubexpIndex("namespace")]))
		Expect(name).To(Equal(groups[re.SubexpIndex("name")]))
	},
	Entry("pod", podRegexp, "/api/v1/namespaces/test/pods/cookie", "test", "cookie"),
	Entry("service", serviceRegexp, "/api/v1/namespaces/test/services/cookie", "test", "cookie"),
	Entry("replicatset", replicasetRegexp, "/apis/apps/v1/namespaces/test/replicasets/cookie", "test", "cookie"),
	Entry("route", routeRegexp, "/apis/route.openshift.io/v1/namespaces/test/routes/cookie", "test", "cookie"),
)

var _ = DescribeTable("list pods",
	func(urlStr string, expectedCount int) {
		objs := []runtimeclient.Object{
			&corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "test",
					Name:      "pod-1",
					Labels: map[string]string{
						"app":  "cookie",
						"with": "chocolate",
					},
				},
			},
			&corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "test",
					Name:      "pod-1",
					Labels: map[string]string{
						"app": "cookie",
					},
				},
			},
		}
		u, err := url.Parse(urlStr)
		Expect(err).NotTo(HaveOccurred())
		Expect(podsRegex.MatchString(u.String())).To(BeTrue())
		logger := logr.New(os.Stdout)
		logger.SetLevel(logr.DebugLevel)
		pods, err := listPods(logger, objs, u)
		Expect(err).NotTo(HaveOccurred())
		Expect(pods.Items).To(HaveLen(expectedCount))
	},
	Entry("single match", "/api/v1/namespaces/test/pods?labelSelector=with%3Dchocolate", 1),
	Entry("multiple matches", "/api/v1/namespaces/test/pods?labelSelector=app%3Dcookie", 2),
	Entry("no match", "/api/v1/namespaces/test/pods?labelSelector=with%3Dnothing", 0),
)

var _ = DescribeTable("list events",
	func(urlStr string, expectedCount int) {
		objs := []runtimeclient.Object{
			&corev1.Event{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "test",
					Name:      "pod-1-event-1",
				},
				InvolvedObject: corev1.ObjectReference{
					Namespace: "test",
					Name:      "pod-1",
				},
			},
			&corev1.Event{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "test",
					Name:      "pod-2-event-1",
				},
				InvolvedObject: corev1.ObjectReference{
					Namespace: "test",
					Name:      "pod-2",
				},
			},
			&corev1.Event{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "test",
					Name:      "pod-2-event-2",
				},
				InvolvedObject: corev1.ObjectReference{
					Namespace: "test",
					Name:      "pod-2",
				},
			},
		}
		u, err := url.Parse(urlStr)
		Expect(err).NotTo(HaveOccurred())
		Expect(eventsRegex.MatchString(u.String())).To(BeTrue())
		logger := logr.New(os.Stdout)
		logger.SetLevel(logr.DebugLevel)
		pods, err := listEvents(logger, objs, u)
		Expect(err).NotTo(HaveOccurred())
		Expect(pods.Items).To(HaveLen(expectedCount))
	},
	Entry("single match", "/api/v1/namespaces/test/events?fieldSelector=involvedObject.namespace%3Dtest%2CinvolvedObject.name%3Dpod-1", 1),
	Entry("multiple matches", "/api/v1/namespaces/test/events?fieldSelector=involvedObject.namespace%3Dtest%2CinvolvedObject.name%3Dpod-2", 2),
	Entry("no match", "/api/v1/namespaces/test/events?fieldSelector=involvedObject.namespace%3Dtest%2CinvolvedObject.name%3Dunknown", 0),
)
