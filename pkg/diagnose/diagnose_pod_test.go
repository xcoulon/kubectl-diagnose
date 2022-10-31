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

	It("should detect ImagePullBackOff", func() {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/pod-image-pull-back-off.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromPod(logger, cfg, "default", "image-pull-back-off")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'ImagePullBackOff': Back-off pulling image "docker.io/unknown:latest"`))
	})

	It("should detect CreateContainerConfigError with logs", func() {
		// given
		logger := logr.New(os.Stdout)
		logger.SetLevel(logr.DebugLevel)
		apiserver, err := NewFakeAPIServer(logger, "resources/pod-create-container-config-error.yaml", "resources/pod-create-container-config-error.logs")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromPod(logger, cfg, "default", "create-container-config-error")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		// container status
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'CreateContainerConfigError': back-off 5m0s restarting failed container=default pod=create-container-config-error`))
		// event
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod events...`))
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-10-31T16:51:05Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Back-off restarting failed container`, time.Since(lastTimestamp).Truncate(time.Second))))
		// logs
		Expect(logger.Output()).To(ContainSubstring(`👀 checking logs in 'default' container...`))
		Expect(logger.Output()).To(ContainSubstring(`🗒  2022/10/31 16:53:38 [emerg] 1#1: mkdir() "/var/lib/nginx/proxy" failed (13: Permission denied)`))
	})

	It("should detect ConfigMap mount error", func() {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/pod-unknown-configmap.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromPod(logger, cfg, "default", "unknown-configmap")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'CreateContainerConfigError': configmap "unknown-configmap" not found`))

	})

	It("should detect Readiness Probe error", func() {
		// given
		logger := logr.New(os.Stdout)
		apiserver, err := NewFakeAPIServer(logger, "resources/pod-readiness-probe-error.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromPod(logger, cfg, "default", "readiness-probe-error")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-10-22T18:48:17Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Readiness probe failed: Get "http://172.17.0.4:80/healthz": dial tcp 172.17.0.4:80: connect: connection refused`, time.Since(lastTimestamp).Truncate(time.Second))))
	})

	It("should detect CrashLoopBackOff error with logs", func() {
		// given
		logger := logr.New(os.Stdout)
		apiserver, err := NewFakeAPIServer(logger, "resources/pod-crash-loop-back-off.yaml", "resources/pod-crash-loop-back-off.logs")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromPod(logger, cfg, "default", "crash-loop-back-off-error")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		// event
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod events...`))
		lastTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2022-10-31T19:08:48Z")
		Expect(logger.Output()).To(ContainSubstring(fmt.Sprintf(`⚡️ %s ago: Back-off restarting failed container`, time.Since(lastTimestamp).Truncate(time.Second))))
		// logs
		Expect(logger.Output()).To(ContainSubstring(`👀 checking logs in 'default' container...`))
		Expect(logger.Output()).To(ContainSubstring(`🗒  Error: loading initial config: loading new config: http app module: start: listening on :80: listen tcp :80: bind: permission denied`))

	})

})
