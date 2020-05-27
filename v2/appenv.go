package appenv

import (
	"context"
)

type ApplicationWithEnvironment interface {
	GetApplicationEnvironments(context.Context) (Result, error)
}

func GetApplicationEnvironments(o ApplicationWithEnvironment, ctx context.Context) (Result, error) {
	if o, ok := o.(ApplicationWithEnvironment); ok {
		return o.GetApplicationEnvironments(ctx)
	}

	return nil, nil
}

func MustGetApplicationEnvironments(o ApplicationWithEnvironment, ctx context.Context) Result {
	result, err := GetApplicationEnvironments(o, ctx)
	if err != nil {
		panic(err)
	}

	return result
}
