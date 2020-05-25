package tests_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"

	. "github.com/holyhope/appenv-generator/tests"
	appenv "github.com/holyhope/appenv-generator/v1"
)

var _ = Describe("Structure", func() {
	var structToTest *Parallel

	Context("With multiple GetApplicationEnvironments taking time", func() {
		sleep := Sleep100{
			SleepDuration: time.Millisecond * 200,
			Values: []v1.EnvVar{
				{
					Name:  "the-key",
					Value: "the-value",
				},
			},
		}

		BeforeEach(func() {
			structToTest = &Parallel{
				FirstSleep:  sleep,
				SecondSleep: sleep,
				ThirdSleep:  sleep,
				FourthSleep: sleep,
				FifthSleep:  sleep,
			}
		})

		It("Should be parallelized", func() {
			start := time.Now()
			result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
			Expect(err).To(Succeed())
			end := time.Now()

			Expect(result.GetEnvs()).To(HaveLen(ParallelCount))
			Expect(end.Sub(start)).Should(BeNumerically("<", sleep.SleepDuration*ParallelCount))
		})

		Describe("Canceling the context", func() {
			It("Should return the right result", func() {
				ctx, cancel := context.WithCancel(context.Background())

				go func() {
					defer GinkgoRecover()

					start := time.Now()
					_, err := appenv.GetApplicationEnvironments(structToTest, ctx)
					Expect(err).To(HaveOccurred())

					end := time.Now()
					Expect(end.Sub(start)).Should(BeNumerically("<", sleep.SleepDuration))
				}()

				cancel()
			})
		})
	})
})
