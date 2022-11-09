package diagnose_test

import (
	"io"
	"os"

	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	. "github.com/xcoulon/kubectl-diagnose/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("diagnose replicasets", func() {

	Context("serviceaccount not found", func() {

		It("should detect from replicaset", func() {
			// given
			logger := logr.New(io.Discard)
			apiserver, err := NewFakeAPIServer(logger, "resources/replicaset-service-account-not-found.yaml")
			Expect(err).NotTo(HaveOccurred())
			cfg := NewConfig(apiserver.URL, "/api")

			// when
			found, err := diagnose.Diagnose(logger, cfg, diagnose.ReplicaSet, "default", "sa-notfound")

			// then
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(logger.Output()).To(ContainSubstring(`👀 checking ReplicaSet 'sa-notfound'...`))
			Expect(logger.Output()).To(ContainSubstring(`👻 replicaset 'sa-notfound' failed to create pods: pods "sa-notfound-" is forbidden: error looking up service account test/sa-notfound: serviceaccount "sa-notfound" not found`))
		})

		It("should detect from route", func() {
			// given
			logger := logr.New(io.Discard)
			apiserver, err := NewFakeAPIServer(logger, "resources/replicaset-service-account-not-found.yaml")
			Expect(err).NotTo(HaveOccurred())
			cfg := NewConfig(apiserver.URL, "/api")

			// when
			found, err := diagnose.Diagnose(logger, cfg, diagnose.Route, "default", "sa-notfound")

			// then
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(logger.Output()).To(ContainSubstring(`👀 checking ReplicaSet 'sa-notfound'...`))
			Expect(logger.Output()).To(ContainSubstring(`👻 replicaset 'sa-notfound' failed to create pods: pods "sa-notfound-" is forbidden: error looking up service account test/sa-notfound: serviceaccount "sa-notfound" not found`))
		})

	})

	It("should detect crash-loop-back-off pod", func() {
		// given
		logger := logr.New(os.Stdout)
		logger.SetLevel(logr.DebugLevel)
		apiserver, err := NewFakeAPIServer(logger, "resources/replicaset-crash-loop-back-off.yaml", "resources/replicaset-crash-loop-back-off.logs")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, diagnose.ReplicaSet, "default", "crash-loop-back-off-8644b7cf9d")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking ReplicaSet 'crash-loop-back-off-8644b7cf9d'...`))
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'crash-loop-back-off-8644b7cf9d-s2fpd'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 containers with unready status: [default]`))
		Expect(logger.Output()).To(ContainSubstring(`👻 container 'default' is waiting with reason 'CrashLoopBackOff': back-off 5m0s restarting failed container=default pod=crash-loop-back-off-8644b7cf9d-s2fpd_xcoulon-2-dev(024449bc-83cf-4251-b39d-3d63ae03d6a2)`))
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod events...`))
		Expect(logger.Output()).To(ContainSubstring(`👀 checking logs in 'default' container...`))
		Expect(logger.Output()).To(ContainSubstring(`🗒  2022/11/09 09:57:27 [emerg] 1#1: mkdir() "/var/lib/nginx/proxy" failed (13: Permission denied)`))
		Expect(logger.Output()).To(ContainSubstring(`🗒  nginx: [emerg] mkdir() "/var/lib/nginx/proxy" failed (13: Permission denied)`))
	})

})
