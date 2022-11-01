package diagnose_test

import (
	"io"

	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	. "github.com/xcoulon/kubectl-diagnose/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("diagnose replicasets", func() {

	It("should detect sa not found from replicaset", func() {
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
		Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'sa-notfound'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 replicaset 'sa-notfound' failed to create pods: pods "sa-notfound-" is forbidden: error looking up service account test/sa-notfound: serviceaccount "sa-notfound" not found`))
	})

	It("should detect sa not found from route", func() {
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
		Expect(logger.Output()).To(ContainSubstring(`👀 checking replicaset 'sa-notfound'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 replicaset 'sa-notfound' failed to create pods: pods "sa-notfound-" is forbidden: error looking up service account test/sa-notfound: serviceaccount "sa-notfound" not found`))
	})

})
