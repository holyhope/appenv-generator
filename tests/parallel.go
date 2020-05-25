package tests

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/holyhope/appenv-generator/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	ParallelCount = 5
)

type Sleep struct {
	SleepDuration time.Duration
	Values        []v1.EnvVar
}

func (s *Sleep) GetApplicationEnvironments(ctx context.Context) (appenv.Result, error) {
	select {
	case <-time.After(s.SleepDuration):
		return appenv.NewResult(s.Values, nil), nil
	case <-ctx.Done():
		return nil, context.Canceled
	}
}

type Parallel struct {
	First  Sleep
	Second Sleep
	Third  Sleep
	Fourth Sleep
	Fifth  Sleep
}

type ErrorOnce struct {
	*sync.Once
}

func (o *ErrorOnce) GetApplicationEnvironments(ctx context.Context) (result appenv.Result, err error) {
	o.Do(func() {
		err = errors.New("expected error for test")
	})
	if err != nil {
		return nil, err
	}

	select {
	case <-time.After(time.Millisecond * 500):
		return nil, nil
	case <-ctx.Done():
		return nil, context.Canceled
	}
}

type ParallelWithError struct {
	First  ErrorOnce
	Second ErrorOnce
	Third  ErrorOnce
	Fourth ErrorOnce
	Fifth  ErrorOnce
}
