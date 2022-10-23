package diagnose_test

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	. "github.com/xcoulon/kubectl-diagnose/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("diagnose pods", func() {

	It("should detect image pull backoff", func() {
		// given
		logger := logr.New(io.Discard)
		apiserver, _, err := NewFakeAPIServer(logger, "resources/pod-image-pull-backoff.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromPod(logger, cfg, "test", "image-pull-backoff")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'ImagePullBackOff': Back-off pulling image "docker.io/unknown:latest"`))
	})

	It("should detect configuration error", func() {
		// given
		logger := logr.New(io.Discard)
		apiserver, _, err := NewFakeAPIServer(logger, "resources/pod-container-config-error.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromPod(logger, cfg, "test", "container-config-error")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'CreateContainerConfigError': container has runAsNonRoot and image will run as root (pod: "container-config-error_test(2513d2ac-91fa-4d90-b8f8-45d9c438d946)", container: default)`))
	})

	It("should detect configmap mount error", func() {
		// given
		logger := logr.New(io.Discard)
		apiserver, _, err := NewFakeAPIServer(logger, "resources/pod-unknown-configmap.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromPod(logger, cfg, "test", "unknown-configmap")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'CreateContainerConfigError': configmap "unknown-configmap" not found`))

	})

	It("should detect readiness probe error", func() {
		// given
		logger := logr.New(os.Stdout)
		apiserver, _, err := NewFakeAPIServer(logger, "resources/pod-readiness-probe-error.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromPod(logger, cfg, "test", "readiness-probe-error")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		lastTimestamp, err := time.Parse("2006-01-02T03:04:05Z", "2022-10-22T08:48:17Z")
		Expect(err).NotTo(HaveOccurred())
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Readiness probe failed: Get "http://172.17.0.4:80/healthz": dial tcp 172.17.0.4:80: connect: connection refused`, time.Since(lastTimestamp).Truncate(time.Second))))
	})

})
