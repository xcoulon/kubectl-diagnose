package test

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
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

func NewFakeAPIServer(logger logr.Logger, filenames ...string) (*httptest.Server, error) {
	allObjs := []runtimeclient.Object{}
	allLogs := map[string]map[string][]string{}
	for _, filename := range filenames {
		switch filepath.Ext(filename) {
		case ".yaml":
			objs, err := parseObjects(filename)
			if err != nil {
				return nil, err
			}
			allObjs = append(allObjs, objs...)
		case ".logs":
			logs, err := parseLogs(filename)
			if err != nil {
				return nil, err
			}
			for p, cl := range logs {
				for c, l := range cl {
					if allLogs[p] == nil {
						allLogs[p] = map[string][]string{}
					}
					allLogs[p][c] = l
				}
			}
		default:
			return nil, fmt.Errorf("unsupported file kind: '%s'", filepath.Ext(filename))
		}
	}
	r := httprouter.New()
	r.GET(`/api/v1/namespaces/:namespace/pods/:name`, newObjectHandler(logger, allObjs, "Pod"))
	r.GET(`/api/v1/namespaces/:namespace/pods`, newPodsHandler(logger, allObjs))
	r.GET(`/api/v1/namespaces/:namespace/pods/:name/log`, newPodLogsHandler(logger, allLogs))
	r.GET(`/apis/route.openshift.io/v1/namespaces/:namespace/routes/:name`, newObjectHandler(logger, allObjs, "Route"))
	r.GET(`/api/v1/namespaces/:namespace/services/:name`, newObjectHandler(logger, allObjs, "Service"))
	r.GET(`/apis/apps/v1/namespaces/:namespace/replicasets/:name`, newObjectHandler(logger, allObjs, "ReplicaSet"))
	r.GET(`/api/v1/namespaces/:namespace/events`, newEventsHandler(logger, allObjs))
	r.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Errorf("no match for request with path='%s' and query='%s' ", r.URL.Path, r.URL.Query().Encode())
		w.WriteHeader(http.StatusNotFound)
	})
	return httptest.NewServer(r), nil
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

func parseLogs(filename string) (map[string]map[string][]string, error) {
	logs := map[string]map[string][]string{}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	data := map[string]interface{}{}
	if err := decoder.Decode(data); err != nil {
		return nil, err
	}
	for p, e := range data {
		logs[p] = map[string][]string{}
		if e, ok := e.(map[string]interface{}); ok {
			for c, l := range e {
				if l, ok := l.(string); ok {
					logs[p][c] = []string{}
					scanner := bufio.NewScanner(strings.NewReader(l))
					scanner.Split(bufio.ScanLines)
					for scanner.Scan() {
						logs[p][c] = append(logs[p][c], scanner.Text())
					}
				}
			}
		}
	}
	return logs, nil
}

// ----------------------------------
// Endpoint Handlers
// ----------------------------------

type NotFoundErr struct {
	msg string
}

func NewNotFoundErr(msg string) error {
	return NotFoundErr{
		msg: msg,
	}
}

func (e NotFoundErr) Error() string {
	return e.msg
}

func newObjectHandler(logger logr.Logger, objs []runtimeclient.Object, kind string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		logger.Infof("handling object at '%s'", r.URL.Path)
		namespace := params.ByName("namespace")
		name := params.ByName("name")
		obj, err := lookupObject(logger, kind, namespace, name, objs)
		if err != nil {
			handleError(logger, w, err)
			return
		}
		handleObject(logger, w, obj)
	}
}

func lookupObject(logger logr.Logger, kind, namespace, name string, objs []runtimeclient.Object) (interface{}, error) {
	logger.Debugf("looking up %s/%s", namespace, name)
	for _, obj := range objs {
		if obj.GetObjectKind().GroupVersionKind().Kind == kind &&
			obj.GetNamespace() == namespace &&
			obj.GetName() == name {
			return obj, nil
		}
	}
	return nil, NewNotFoundErr(fmt.Sprintf("no match for %s/%s (missing resource?)", namespace, name))
}

func newPodLogsHandler(logger logr.Logger, logs map[string]map[string][]string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		namespace := params.ByName("namespace")
		pod := params.ByName("name")
		container := r.URL.Query().Get("container")
		logger.Debugf("fetching logs for container '%s' of pod '%s' in namespace '%s'", container, pod, namespace)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		output := strings.Join(logs[pod][container], "\n")
		w.Write([]byte(output)) //nolint: errcheck
	}
}

func newPodsHandler(logger logr.Logger, objs []runtimeclient.Object) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		namespace := params.ByName("namespace")
		labelSelector := r.URL.Query().Get("labelSelector")
		logger.Debugf("listing pods in %s with labels %s", namespace, labelSelector)
		s, err := labels.Parse(labelSelector)
		if err != nil {
			handleError(logger, w, err)
			return
		}
		pods := &corev1.PodList{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "PodList",
			},
			Items: []corev1.Pod{},
		}
		for _, obj := range objs {
			if obj, ok := obj.(*corev1.Pod); ok &&
				obj.GetNamespace() == namespace &&
				s.Matches(labels.Set(obj.GetLabels())) {
				pods.Items = append(pods.Items, *obj)
			}
		}
		handleObject(logger, w, pods)
	}
}

func newEventsHandler(logger logr.Logger, objs []runtimeclient.Object) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		namespace := params.ByName("namespace")
		fieldSelector := r.URL.Query().Get("fieldSelector")
		logger.Debugf("listing events in %s with fields %s", namespace, fieldSelector)
		s, err := fields.ParseSelector(fieldSelector)
		if err != nil {
			handleError(logger, w, err)
			return
		}
		events := &corev1.EventList{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "EventList",
			},
			Items: []corev1.Event{},
		}
		for _, obj := range objs {
			if obj, ok := obj.(*corev1.Event); ok &&
				obj.GetNamespace() == namespace &&
				s.Matches(fields.Set(map[string]string{
					"involvedObject.kind":      "Pod",
					"involvedObject.namespace": obj.InvolvedObject.Namespace,
					"involvedObject.name":      obj.InvolvedObject.Name,
				})) {
				events.Items = append(events.Items, *obj)
			}
		}
		handleObject(logger, w, events)
	}
}

func handleObject(logger logr.Logger, w http.ResponseWriter, obj interface{}) {
	output, err := json.Marshal(obj)
	if err != nil {
		logger.Errorf(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(output) //nolint: errcheck
}

func handleError(logger logr.Logger, w http.ResponseWriter, err error) {
	logger.Errorf(err.Error())
	switch {
	case errors.Is(err, NotFoundErr{}):
		w.WriteHeader(http.StatusNotFound)

	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
