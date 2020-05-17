package appenv

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

type ApplicationWithEnvironment interface {
	GetApplicationEnvironments(context.Context) []v1.EnvVar
}
