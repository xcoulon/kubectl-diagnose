package test

import (
	"net/url"
	"os"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

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
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "pod-1",
					Labels: map[string]string{
						"app":  "cookie",
						"with": "chocolate",
					},
				},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
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
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "pod-1-event-1",
				},
				InvolvedObject: corev1.ObjectReference{
					Namespace: "test",
					Name:      "pod-1",
				},
			},
			&corev1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "pod-2-event-1",
				},
				InvolvedObject: corev1.ObjectReference{
					Namespace: "test",
					Name:      "pod-2",
				},
			},
			&corev1.Event{
				ObjectMeta: metav1.ObjectMeta{
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
		pods, err := listEvents(logger, objs, u)
		Expect(err).NotTo(HaveOccurred())
		Expect(pods.Items).To(HaveLen(expectedCount))
	},
	Entry("single match", "/api/v1/namespaces/test/events?fieldSelector=involvedObject.namespace%3Dtest%2CinvolvedObject.name%3Dpod-1", 1),
	Entry("multiple matches", "/api/v1/namespaces/test/events?fieldSelector=involvedObject.namespace%3Dtest%2CinvolvedObject.name%3Dpod-2", 2),
	Entry("no match", "/api/v1/namespaces/test/events?fieldSelector=involvedObject.namespace%3Dtest%2CinvolvedObject.name%3Dunknown", 0),
)
