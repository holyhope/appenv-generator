package appenv

import (
	"context"
	"sync"

	v1 "k8s.io/api/core/v1"

	"github.com/holyhope/appenv-generator"
	appenvv1 "github.com/holyhope/appenv-generator/v1"
)

type Result interface {
	appenv.Result
	appenvv1.Result

	GetVolumes() []v1.Volume
	GetVolumesMount() []v1.VolumeMount
}

func NewResult(envs []v1.EnvVar, envsFrom []v1.EnvFromSource, volumes []v1.Volume, volumesMount []v1.VolumeMount) *result {
	return &result{
		envs:         envs,
		envsFrom:     envsFrom,
		volumes:      volumes,
		volumesMount: volumesMount,
	}
}

type result struct {
	envs     []v1.EnvVar
	lockEnvs sync.Mutex

	envsFrom     []v1.EnvFromSource
	lockEnvsFrom sync.Mutex

	volumes     []v1.Volume
	lockVolumes sync.Mutex

	volumesMount     []v1.VolumeMount
	lockVolumesMount sync.Mutex
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

func (r *result) GetVolumes() []v1.Volume {
	return r.volumes
}

func (r *result) AddVolumes(volumes ...v1.Volume) {
	r.lockVolumes.Lock()
	defer r.lockVolumes.Unlock()

	r.volumes = append(r.volumes, volumes...)
}

func (r *result) GetVolumesMount() []v1.VolumeMount {
	return r.volumesMount
}

func (r *result) AddVolumesMount(volumesMount ...v1.VolumeMount) {
	r.lockVolumesMount.Lock()
	defer r.lockVolumesMount.Unlock()

	r.volumesMount = append(r.volumesMount, volumesMount...)
}

func (r *result) Add(_ context.Context, res appenv.Result) error {
	if res == nil {
		return nil
	}

	var wg sync.WaitGroup

	if resultV2, ok := res.(*result); ok {
		wg.Add(1)
		go func() {
			defer wg.Done()

			r.AddVolumes(resultV2.GetVolumes()...)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			r.AddVolumesMount(resultV2.GetVolumesMount()...)
		}()
	}

	if resultV1, ok := res.(appenvv1.Result); ok {
		wg.Add(1)
		go func() {
			defer wg.Done()

			r.AddEnvs(resultV1.GetEnvs()...)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			r.AddEnvsFrom(resultV1.GetEnvsFrom()...)
		}()
	} else {
		return appenv.NewVersionError(res)
	}

	wg.Wait()

	return nil
}
