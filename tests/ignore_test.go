package tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/holyhope/appenv-generator/tests"
	appenv "github.com/holyhope/appenv-generator/v1"
)

var _ = Describe("Structure", func() {
	var structToTest interface{}

	Context("With custom GetApplicationEnvironments implementation", func() {
		BeforeEach(func() {
			structToTest = &Ignore{
				SimpleString: "soSimple",
				SimpleFloat:  1965,
			}
		})

		It("Should not implements ApplicationWithEnvironment interface", func() {
			_, ok := structToTest.(appenv.ApplicationWithEnvironment)
			Expect(ok).To(BeFalse())
		})
	})
})
