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
	var structToTest *Volume

	Context("With custom GetApplicationEnvironments implementation", func() {
		BeforeEach(func() {
			structToTest = &Volume{
				Secret:      "the-secret",
				Certificate: "certificate-name",
				FullSecret:  "full-secret",
			}
		})

		It("Should return the right result", func() {
			result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
			Expect(err).To(Succeed())
			Expect(result.GetEnvs()).To(ConsistOf(v1.EnvVar{
				Name:  "FULL_SECRET_MOUNT_LOCATION",
				Value: "/the/full-secret-mount",
			}))
			Expect(result.GetEnvsFrom()).To(BeEmpty())
			Expect(result.GetVolumes()).To(ConsistOf(
				v1.Volume{
					Name: "the-secret",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "the-secret",
						},
					},
				},
				v1.Volume{
					Name: "the-certificate",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "certificate-name",
							Items: []v1.KeyToPath{
								{
									Key:  "tls.key",
									Path: "private_key.pem",
								},
							},
						},
					},
				},
				v1.Volume{
					Name: "full-secret",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "full-secret",
							Items: []v1.KeyToPath{
								{
									Key:  "tls.key",
									Path: "private_key.pem",
								},
							},
						},
					},
				},
			))
			Expect(result.GetVolumesMount()).To(ConsistOf(
				v1.VolumeMount{
					Name:      "the-secret",
					MountPath: "/the/secret",
				},
				v1.VolumeMount{
					Name:      "the-certificate",
					MountPath: "/etc/certificates",
				},
				v1.VolumeMount{
					Name:      "full-secret",
					MountPath: "/the/full-secret-mount",
				},
			))
		})
	})
})
