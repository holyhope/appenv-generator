package tests_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"

	. "github.com/holyhope/appenv-generator/tests"
	appenv "github.com/holyhope/appenv-generator/v1"
)

var _ = Describe("Structure", func() {
	var structToTest interface{}

	Context("With custom GetApplicationEnvironments implementation", func() {
		BeforeEach(func() {
			structToTest = &Custom{}
		})

		It("Should return the right result", func() {
			envs, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
			Expect(err).To(Succeed())
			Expect(envs).Should(ConsistOf(v1.EnvVar{
				Name:  "custom-key",
				Value: "custom-value",
			}))
		})
	})
})
