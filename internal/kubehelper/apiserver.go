package kubehelper

import (
	"github.com/blang/semver"
)

type KubeAPI struct {
}

// ApiVersion returns the version of the remote kubernetes API server
func (k *KubeAPI) Version() (semver.Version, error) {
	client, err := createKubeClient()
	if err != nil {
		return semver.Version{}, err
	}

	v, err := client.DiscoveryClient.ServerVersion()
	if err != nil {
		return semver.Version{}, err
	}
	return semver.ParseTolerant(v.GitVersion)
}
