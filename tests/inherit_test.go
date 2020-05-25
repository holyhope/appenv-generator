package tests_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/holyhope/appenv-generator/tests"
	appenv "github.com/holyhope/appenv-generator/v1"
)

var _ = Describe("Structure", func() {
	var structToTest *Inherit

	Context("With underlying fields", func() {
		var children1 *SimpleTest
		var children2 *SimpleTest2

		BeforeEach(func() {
			children1 = &SimpleTest{
				SimpleString: "the first value",
			}
			value := "a second value"
			children2 = &SimpleTest2{
				SimpleString2: &value,
			}
			structToTest = &Inherit{
				SimpleTest:  *children1,
				SimpleTest2: children2,
			}
		})

		It("Should return the merged result", func() {
			result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
			Expect(err).To(Succeed())

			envs := result.GetEnvs()
			Expect(envs).ToNot(BeEmpty())

			result1, _ := appenv.GetApplicationEnvironments(children1, context.TODO())
			result2, _ := appenv.GetApplicationEnvironments(children2, context.TODO())
			Expect(envs).Should(ConsistOf(append(result1.GetEnvs(), result2.GetEnvs()...)))
		})
	})
})
