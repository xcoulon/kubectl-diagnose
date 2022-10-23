package diagnose_test

import (
	"os"

	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	. "github.com/xcoulon/kubectl-diagnose/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("diagnose services", func() {

	It("should not detect errors when all good", func() {
		// given
		logger := logr.New(os.Stdout)
		logger.SetLevel(logr.DebugLevel)
		// TODO: service should have multiple ports
		apiserver, _, err := NewFakeAPIServer(logger, "resources/all-good.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromService(logger, cfg, "test", "all-good")
		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeFalse())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'all-good' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`☑️ found matching target port 'http' (8080) in container 'default' of pod 'all-good-785d8bcc5f-g92mn'`))
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'all-good-785d8bcc5f-g92mn'...`))
		Expect(logger.Output()).To(ContainSubstring(`☑️ found matching target port 'http' (8080) in container 'default' of pod 'all-good-785d8bcc5f-x85p2'`))
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'all-good-785d8bcc5f-x85p2'...`))
	})

	It("should detect no matching pods", func() {
		// given
		logger := logr.New(os.Stdout)
		logger.SetLevel(logr.DebugLevel)
		apiserver, _, err := NewFakeAPIServer(logger, "resources/service-no-matching-pods.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromService(logger, cfg, "test", "service-no-matching-pods")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'service-no-matching-pods' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no pods matching label selector 'app=invalid' found in namespace 'test'`))
		Expect(logger.Output()).To(ContainSubstring(`💡 you may want to:`))
		Expect(logger.Output()).To(ContainSubstring(` - check the 'service.spec.selector' value`))
		Expect(logger.Output()).To(ContainSubstring(` - make sure that the expected pods exists`))
	})

	It("should detect invalid target port as string", func() {
		// given
		logger := logr.New(os.Stdout)
		logger.SetLevel(logr.DebugLevel)
		apiserver, _, err := NewFakeAPIServer(logger, "resources/service-invalid-target-port.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromService(logger, cfg, "test", "service-invalid-target-port")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'service-invalid-target-port' in namespace 'test'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no container with matching target port 'https' in pod 'service-invalid-target-port-68968cf979-wtpcp'`))
	})

})
