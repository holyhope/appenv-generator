package appenv

import (
	"context"
	"sync"

	"github.com/holyhope/appenv-generator"
	v1 "k8s.io/api/core/v1"
)

type Result interface {
	appenv.Result

	GetEnvs() []v1.EnvVar
	GetEnvsFrom() []v1.EnvFromSource
}

func NewResult(envs []v1.EnvVar, envsFrom []v1.EnvFromSource) *result {
	return &result{
		envs:     envs,
		envsFrom: envsFrom,
	}
}

type result struct {
	envs     []v1.EnvVar
	lockEnvs sync.Mutex

	envsFrom     []v1.EnvFromSource
	lockEnvsFrom sync.Mutex
}

func (r *result) GetEnvs() []v1.EnvVar {
	return r.envs
}

func (r *result) AddEnvs(envs ...v1.EnvVar) {
	r.lockEnvs.Lock()
	defer r.lockEnvs.Unlock()

	r.envs = append(r.envs, envs...)
}

func (r *result) GetEnvsFrom() []v1.EnvFromSource {
	return r.envsFrom
}

func (r *result) AddEnvsFrom(envs ...v1.EnvFromSource) {
	r.lockEnvsFrom.Lock()
	defer r.lockEnvsFrom.Unlock()

	r.envsFrom = append(r.envsFrom, envs...)
}

func (r *result) Add(_ context.Context, res appenv.Result) error {
	if res == nil {
		return nil
	}

	resultV1, ok := res.(*result)
	if !ok {
		return appenv.NewVersionError(res)
	}

	r.AddEnvs(resultV1.GetEnvs()...)
	r.AddEnvsFrom(resultV1.GetEnvsFrom()...)

	return nil
}
