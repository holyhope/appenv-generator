package tests

import (
	"context"
	"time"

	"github.com/holyhope/appenv-generator/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	ParallelCount = 5
)

type Sleep100 struct {
	SleepDuration time.Duration
	Values        []v1.EnvVar
}

func (s *Sleep100) GetApplicationEnvironments(context.Context) (appenv.Result, error) {
	time.Sleep(s.SleepDuration)
	return appenv.NewResult(s.Values, nil), nil
}

type Parallel struct {
	FirstSleep  Sleep100
	SecondSleep Sleep100
	ThirdSleep  Sleep100
	FourthSleep Sleep100
	FifthSleep  Sleep100
}
