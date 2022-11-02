package diagnose_test

import (
	"io"

	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	. "github.com/xcoulon/kubectl-diagnose/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("diagnose services", func() {

	It("should not detect errors when all good", func() {
		// given
		logger := logr.New(io.Discard)
		// TODO: service should have multiple ports
		apiserver, err := NewFakeAPIServer(logger, "resources/all-good.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, "svc", "default", "all-good")
		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeFalse())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'all-good' in namespace 'default'...`))
		Expect(logger.Output()).To(ContainSubstring(`☑️ found matching target port 'http' (8080) in container 'default' of pod 'all-good-785d8bcc5f-g92mn'`))
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'all-good-785d8bcc5f-g92mn'...`))
		Expect(logger.Output()).To(ContainSubstring(`☑️ found matching target port 'http' (8080) in container 'default' of pod 'all-good-785d8bcc5f-x85p2'`))
		Expect(logger.Output()).To(ContainSubstring(`👀 checking pod 'all-good-785d8bcc5f-x85p2'...`))
	})

	It("should detect no matching pods", func() {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/service-no-matching-pods.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, "svc", "default", "service-no-matching-pods")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'service-no-matching-pods' in namespace 'default'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no pods matching label selector 'app=invalid' found in namespace 'default'`))
		Expect(logger.Output()).To(ContainSubstring(`💡 you may want to verify that the pods exist and their labels match 'app=invalid'`))
	})

	It("should detect invalid target port as string", func() {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/service-invalid-target-port-str.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, "svc", "default", "service-invalid-target-port-str")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'service-invalid-target-port-str' in namespace 'default'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no container with port 'https' in pod 'service-invalid-target-port-str'`))
	})

	It("should detect invalid target port as int", func() {
		// given
		logger := logr.New(io.Discard)
		apiserver, err := NewFakeAPIServer(logger, "resources/service-invalid-target-port-int.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.Diagnose(logger, cfg, "svc", "default", "service-invalid-target-port-int")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking service 'service-invalid-target-port-int' in namespace 'default'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 no container with port '8443' in pod 'service-invalid-target-port-int'`))
	})

})
