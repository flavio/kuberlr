package kubehelper

import (
	"github.com/blang/semver"
)

// ApiVersion returns the version of the remote kubernetes API server
func ApiVersion() (semver.Version, error) {
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
