package tests_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"

	. "github.com/holyhope/appenv-generator/tests"
	"github.com/holyhope/appenv-generator/v1"
)

var _ = Describe("Structure", func() {
	var structToTest *Parallel

	Context("With multiple GetApplicationEnvironments taking time", func() {
		sleep := Sleep{
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
				First:  sleep,
				Second: sleep,
				Third:  sleep,
				Fourth: sleep,
				Fifth:  sleep,
			}
		})

		It("Should be optimized", func() {
			start := time.Now()
			result, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
			Expect(err).To(Succeed())
			end := time.Now()

			Expect(result.GetEnvs()).To(HaveLen(ParallelCount))
			Expect(result.GetEnvsFrom()).To(BeEmpty())

			Expect(end.Sub(start)).To(BeNumerically("<", sleep.SleepDuration*ParallelCount))
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
					Expect(end.Sub(start)).To(BeNumerically("<", sleep.SleepDuration))
				}()

				cancel()
			})
		})

		Describe("Returning 1 error", func() {
			var structToTest *ParallelWithError

			BeforeEach(func() {
				once := sync.Once{}
				structToTest = &ParallelWithError{
					First:  ErrorOnce{Once: &once},
					Second: ErrorOnce{Once: &once},
					Third:  ErrorOnce{Once: &once},
					Fourth: ErrorOnce{Once: &once},
					Fifth:  ErrorOnce{Once: &once},
				}
			})

			It("Should stop other calls", func() {
				start := time.Now()
				_, err := appenv.GetApplicationEnvironments(structToTest, context.TODO())
				Expect(err).To(HaveOccurred())

				end := time.Now()
				Expect(end.Sub(start)).To(BeNumerically("<", time.Millisecond*500))
			})
		})
	})
})
