package appenv

import (
	"context"
)

type Result interface {
	Add(context.Context, Result) error
}
