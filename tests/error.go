package tests

import (
	"context"
	"errors"

	v1 "k8s.io/api/core/v1"
)

type Error struct{}

func (c *Error) GetApplicationEnvironments(context.Context) ([]v1.EnvVar, error) {
	return nil, errors.New("expected error")
}
