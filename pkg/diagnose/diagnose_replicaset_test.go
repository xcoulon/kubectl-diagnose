package diagnose_test

import (
	"os"

	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
	. "github.com/xcoulon/kubectl-diagnose/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("diagnose replicasets", func() {

	It("should detect sa not found", func() {
		// given
		logger := logr.New(os.Stdout)
		logger.SetLevel(logr.DebugLevel)
		apiserver, _, err := NewFakeAPIServer(logger, "resources/replicaset-service-account-not-found.yaml")
		Expect(err).NotTo(HaveOccurred())
		cfg := NewConfig(apiserver.URL, "/api")

		// when
		found, err := diagnose.DiagnoseFromReplicaSet(logger, cfg, "test", "sa-notfound")

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(logger.Output()).To(ContainSubstring(`👀 checking ReplicaSet 'sa-notfound'...`))
		Expect(logger.Output()).To(ContainSubstring(`👻 replicaset 'sa-notfound' failed to create pods: pods "sa-notfound-" is forbidden: error looking up service account test/sa-notfound: serviceaccount "sa-notfound" not found`))
	})

})
