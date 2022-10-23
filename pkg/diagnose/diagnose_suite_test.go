package diagnose_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDiagnose(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Diagnose Suite")
}
