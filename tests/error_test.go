package tests_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/holyhope/appenv-generator/tests"
	"github.com/holyhope/appenv-generator/v2"
)

var _ = Describe("Structure", func() {
	var structToTest *Error

	Context("With GetApplicationEnvironments returning an error", func() {
		BeforeEach(func() {
			structToTest = &Error{}
		})

		It("Should propagate the error", func() {
			_, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
			Expect(err).To(HaveOccurred())

			Expect(func() {
				appenv.MustGetApplicationEnvironments(structToTest, context.TODO())
			}).To(Panic())
		})
	})
})
