package tests_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/holyhope/appenv-generator/tests"
	"github.com/holyhope/appenv-generator/v1"
)

var _ = Describe("Structure", func() {
	Context("With underlying fields", func() {
		var structToTest *Inherit

		var child1 *SimpleTest
		var child2 *SimpleTest2

		BeforeEach(func() {
			child1 = &SimpleTest{
				SimpleString: "the first value",
			}
			value := "a second value"
			child2 = &SimpleTest2{
				SimpleString2: &value,
			}
			structToTest = &Inherit{
				SimpleTest:  *child1,
				SimpleTest2: child2,
			}
		})

		It("Should return the merged result", func() {
			result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
			Expect(err).To(Succeed())

			result1 := appenv.MustGetApplicationEnvironments(child1, context.TODO())
			result2 := appenv.MustGetApplicationEnvironments(child2, context.TODO())
			Expect(result.GetEnvs()).To(ConsistOf(append(result1.GetEnvs(), result2.GetEnvs()...)))
			Expect(result.GetEnvsFrom()).To(ConsistOf(append(result1.GetEnvsFrom(), result2.GetEnvsFrom()...)))
		})
	})

	Context("With underlying fields", func() {
		var structToTest *InheritFromField

		var child FromField

		BeforeEach(func() {
			child = FromField{
				SecretName:     "single-secret",
				FullSecretName: "full-secrets",
				ConfigName:     "config",
				FullConfigName: "config",
			}
			structToTest = &InheritFromField{
				FromField: child,
			}
		})

		It("Should return the merged result", func() {
			result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
			Expect(err).To(Succeed())

			envs := result.GetEnvs()
			Expect(envs).ToNot(BeEmpty())

			childResult := appenv.MustGetApplicationEnvironments(&child, context.TODO())
			Expect(result.GetEnvs()).To(ConsistOf(childResult.GetEnvs()))
			Expect(result.GetEnvsFrom()).To(ConsistOf(childResult.GetEnvsFrom()))
		})
	})
})
