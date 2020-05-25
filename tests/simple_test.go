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
	Context("With basic typed fields", func() {
		var structToTest *SimpleTest

		BeforeEach(func() {
			structToTest = &SimpleTest{
				SimpleString: "soSimple",
				SimpleFloat:  1965.07,
				SimpleInt:    10,
			}
		})

		It("Should return the right result", func() {
			result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
			Expect(err).To(Succeed())
			Expect(result.GetEnvs()).Should(ConsistOf(v1.EnvVar{
				Name:  "SIMPLE_STRING",
				Value: "soSimple",
			}, v1.EnvVar{
				Name:  "SIMPLE_FLOAT",
				Value: "1965.07",
			}, v1.EnvVar{
				Name:  "SIMPLE_INT",
				Value: "10",
			}, v1.EnvVar{
				Name:  "SIMPLE_BOOL",
				Value: "false",
			}))
		})
	})

	Context("with pointer fields", func() {
		var structToTest *SimpleTest2

		Context("With nil value", func() {
			BeforeEach(func() {
				structToTest = &SimpleTest2{}
			})

			It("Should not return the result", func() {
				result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
				Expect(err).To(Succeed())
				Expect(result.GetEnvs()).Should(BeEmpty())
			})
		})

		Context("with not nil value", func() {
			BeforeEach(func() {
				value := "soSimple2"
				structToTest = &SimpleTest2{
					SimpleString2: &value,
				}
			})

			It("Should return the right result", func() {
				result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
				Expect(err).To(Succeed())
				Expect(result.GetEnvs()).Should(ConsistOf(v1.EnvVar{
					Name:  "SIMPLE_STRING2",
					Value: "soSimple2",
				}))
			})
		})
	})
})
