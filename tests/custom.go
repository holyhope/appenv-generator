package tests

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

type Custom struct{}

func (c *Custom) GetApplicationEnvironments(context.Context) ([]v1.EnvVar, error) {
	return []v1.EnvVar{
		{
			Name:  "custom-key",
			Value: "custom-value",
		},
	}, nil
}

type CustomInherit struct {
	Custom Custom `json:"custom"`
}
