package tests_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"

	. "github.com/holyhope/appenv-generator/tests"
	"github.com/holyhope/appenv-generator/v2"
)

var _ = Describe("Structure", func() {
	Context("With all field ignored", func() {
		var structToTest interface{}

		BeforeEach(func() {
			structToTest = &IgnoreAllField{
				SimpleString: "soSimple",
			}
		})

		It("Should not implements ApplicationWithEnvironment", func() {
			_, ok := structToTest.(appenv.ApplicationWithEnvironment)
			Expect(ok).To(BeFalse())
		})
	})

	Context("With some field ignored", func() {
		var structToTest *Ignore

		BeforeEach(func() {
			structToTest = &Ignore{
				SimpleString: "soSimple",
				SimpleFloat:  1965,
			}
		})

		It("Should return right fields", func() {
			result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
			Expect(err).ToNot(HaveOccurred())
			Expect(result.GetEnvs()).To(ConsistOf(v1.EnvVar{
				Name:  "SIMPLE_STRING",
				Value: "soSimple",
			}))
		})
	})
})
