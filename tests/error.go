package tests

import (
	"context"
	"errors"

	"github.com/holyhope/appenv-generator/v1"
)

type Error struct{}

func (c *Error) GetApplicationEnvironments(context.Context) (appenv.Result, error) {
	return nil, errors.New("expected error")
}
