package diagnose_test

import (
	"os"

	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	. "github.com/xcoulon/kubectl-diagnose/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("diagnose routes", func() {

	It("should detect missing target service", func() {
		// given
		logger := logr.New(os.Stdout)
		logger.SetLevel(logr.DebugLevel)
		apiserver, err := NewFakeAPIServer(logger, "resources/route-missing-target-service.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromRoute(logger, cfg, "default", "missing-target-service")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'missing-target-service' in namespace 'default'`))
		Expect(logger.Output()).To(ContainSubstring("👻 unable to find service 'missing'"))
	})

	It("should detect invalid target port (as string)", func() {
		// given
		logger := logr.New(os.Stdout)
		logger.SetLevel(logr.DebugLevel)
		apiserver, err := NewFakeAPIServer(logger, "resources/route-invalid-target-port-str.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromRoute(logger, cfg, "default", "invalid-target-port-str")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'invalid-target-port-str' in namespace 'default'`))
		Expect(logger.Output()).To(ContainSubstring("👻 route target port 'https' is not defined in service 'invalid-target-port-str'"))
	})

	It("should detect invalid target port (as int)", func() {
		// given
		logger := logr.New(os.Stdout)
		logger.SetLevel(logr.DebugLevel)
		apiserver, err := NewFakeAPIServer(logger, "resources/route-invalid-target-port-int.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromRoute(logger, cfg, "default", "invalid-target-port-int")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking route 'invalid-target-port-int' in namespace 'default'`))
		Expect(logger.Output()).To(ContainSubstring("👻 route target port '8443' is not defined in service 'invalid-target-port-int'"))
	})
})
