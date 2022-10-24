package test

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
	// Kubernetes Pods and Services
	corev1.AddToScheme(scheme.Scheme) //nolint:errcheck
	// Kubernetes ReplicaSets
	appsv1.AddToScheme(scheme.Scheme) //nolint:errcheck
	// OpenShift Routes
	routev1.AddToScheme(scheme.Scheme) //nolint:errcheck
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
			output, _ := json.Marshal(obj) //nolint: errchkjson
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(output) //nolint: errcheck
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
		if err := decoder.Decode(&value); errors.Is(err, io.EOF) {
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
			// force namespace to `default` here (so it works out-of-the-box on a vanilla Kubernetes cluster)
			obj.SetNamespace("default")
			objs = append(objs, obj)
		}
	}
	return objs, nil
}
