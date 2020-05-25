package tests_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"

	. "github.com/holyhope/appenv-generator/tests"
	"github.com/holyhope/appenv-generator/v1"
)

var _ = Describe("Structure", func() {
	var structToTest *Custom

	Context("With custom GetApplicationEnvironments implementation", func() {
		BeforeEach(func() {
			structToTest = &Custom{}
		})

		It("Should return the right result", func() {
			result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
			Expect(err).To(Succeed())
			Expect(result.GetEnvs()).Should(ConsistOf(v1.EnvVar{
				Name:  "custom-key",
				Value: "custom-value",
			}))
			Expect(result.GetEnvsFrom()).Should(BeEmpty())
		})
	})
})
