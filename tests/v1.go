package tests

import (
	"context"

	"github.com/holyhope/appenv-generator/v1"
	v1 "k8s.io/api/core/v1"
)

type WithVersion1 struct {
	FieldV1 CustomV1 `json:"fieldV1"`
}

type CustomV1 struct{}

func (c *CustomV1) GetApplicationEnvironments(context.Context) (appenv.Result, error) {
	return appenv.NewResult([]v1.EnvVar{
		{
			Name:  "custom-key",
			Value: "custom-value",
		},
	}, nil), nil
}
