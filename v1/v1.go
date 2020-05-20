package appenv

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

type ApplicationWithEnvironment interface {
	GetApplicationEnvironments(context.Context) ([]v1.EnvVar, error)
}

func GetApplicationEnvironments(o interface{}, ctx context.Context) ([]v1.EnvVar, error) {
	if o, ok := o.(ApplicationWithEnvironment); ok {
		return o.GetApplicationEnvironments(ctx)
	}

	return nil, nil
}

func MustGetApplicationEnvironments(o interface{}, ctx context.Context) []v1.EnvVar {
	if o, ok := o.(ApplicationWithEnvironment); ok {
		envs, err := o.GetApplicationEnvironments(ctx)
		if err != nil {
			panic(err)
		}

		return envs
	}

	return nil
}
