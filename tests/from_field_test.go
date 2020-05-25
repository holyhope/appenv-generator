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
	Context("Without pointer", func() {
		var structToTest *FromField

		BeforeEach(func() {
			structToTest = &FromField{
				SecretName:     "single-secret",
				FullSecretName: "full-secrets",
				ConfigName:     "config",
				FullConfigName: "config",
			}
		})

		It("Should return the right result", func() {
			result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
			Expect(err).To(Succeed())
			Expect(result.GetEnvs()).To(ConsistOf(v1.EnvVar{
				Name: "THE_SECRET",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "single-secret",
						},
						Key: "the-secret-entry",
					},
				},
			}, v1.EnvVar{
				Name: "THE_CONFIG",
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "config",
						},
						Key: "the-config-entry",
					},
				},
			}))
			Expect(result.GetEnvsFrom()).To(ConsistOf(v1.EnvFromSource{
				SecretRef: &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "full-secrets",
					},
				},
			}, v1.EnvFromSource{
				ConfigMapRef: &v1.ConfigMapEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "config",
					},
				},
			}))
		})
	})

	Context("with pointer fields", func() {
		var structToTest *FromField2

		Context("With nil values", func() {
			BeforeEach(func() {
				structToTest = &FromField2{}
			})

			It("Should not return the right result", func() {
				result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
				Expect(err).To(Succeed())
				Expect(result.GetEnvs()).To(BeEmpty())
			})
		})

		Context("with not nil value", func() {
			BeforeEach(func() {
				value := "single-secret"
				value2 := "config"
				structToTest = &FromField2{
					SecretName:     &value,
					FullConfigName: &value2,
				}
			})

			It("Should return the right result", func() {
				result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
				Expect(err).To(Succeed())
				Expect(result.GetEnvs()).To(ConsistOf(v1.EnvVar{
					Name: "THE_SECRET",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "single-secret",
							},
							Key: "the-secret-entry",
						},
					},
				}))
				Expect(result.GetEnvsFrom()).To(ConsistOf(v1.EnvFromSource{
					ConfigMapRef: &v1.ConfigMapEnvSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "config",
						},
					},
				}))
			})
		})
	})
})
