package kubehelper

import (
	"k8s.io/apimachinery/pkg/version"
)

// ApiVersion returns the version of the remote kubernetes API server
func ApiVersion() (*version.Info, error) {
	client, err := createKubeClient()
	if err != nil {
		return nil, err
	}

	return client.DiscoveryClient.ServerVersion()
}
