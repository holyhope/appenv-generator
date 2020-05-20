package tests

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
)

const (
	ParallelCount = 5
)

type Sleep100 struct {
	SleepDuration time.Duration
}

func (s *Sleep100) GetApplicationEnvironments(context.Context) ([]v1.EnvVar, error) {
	time.Sleep(s.SleepDuration)
	return nil, nil
}

type Parallel struct {
	FirstSleep  Sleep100
	SecondSleep Sleep100
	ThirdSleep  Sleep100
	FourthSleep Sleep100
	FifthSleep  Sleep100
}
