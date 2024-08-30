package kubehelper

import (
	"github.com/blang/semver/v4"
)

// KubeAPI helps interactions with kubernetes API server.
type KubeAPI struct {
}

// Version returns the version of the remote kubernetes API server.
func (k *KubeAPI) Version(timeout int64) (semver.Version, error) {
	client, err := createKubeClient(timeout)
	if err != nil {
		return semver.Version{}, err
	}

	v, err := client.DiscoveryClient.ServerVersion()
	if err != nil {
		return semver.Version{}, err
	}
	return semver.ParseTolerant(v.GitVersion)
}
