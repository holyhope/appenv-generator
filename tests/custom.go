package tests

import (
	"context"

	"github.com/holyhope/appenv-generator/v1"
	v1 "k8s.io/api/core/v1"
)

type Custom struct{}

func (c *Custom) GetApplicationEnvironments(context.Context) (appenv.Result, error) {
	return appenv.NewResult([]v1.EnvVar{
		{
			Name:  "custom-key",
			Value: "custom-value",
		},
	}, nil), nil
}
