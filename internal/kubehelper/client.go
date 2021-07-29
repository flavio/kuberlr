package kubehelper

import (
	"os"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func createKubeClient(timeout int64) (*kubernetes.Clientset, error) {
	var cliKubeconfig string
	for i := 1; i < len(os.Args); i++ {
		if i+1 < len(os.Args) && os.Args[i] == "--kubeconfig" {
			cliKubeconfig = os.Args[i+1]
			// don't break here; in case there are multiple --kubeconfig options,
			// the last one takes precedence
			continue
		}
		if strings.HasPrefix(os.Args[i], "--kubeconfig=") {
			cliKubeconfig = strings.TrimPrefix(os.Args[i], "--kubeconfig=")
			continue
		}
		if os.Args[i] == "--" {
			break
		}
	}

	var restConfig *restclient.Config
	var err error

	if cliKubeconfig != "" {
		// give precedence to --kubeconfig flag
		restConfig, err = clientcmd.BuildConfigFromFlags("", cliKubeconfig)
	} else {
		// Let the NewDefaultClientConfigLoadingRules do the heavy lifting like
		// parsing the KUBECONFIG value
		// TIL: it's possible to specify multiple kubeconfig files via KUBECONFIG
		// For example: `KUBECONFIG=~/cluster1.yaml:~/cluster2.yaml`
		// See https://github.com/kubernetes/kubernetes/issues/46381#issuecomment-303926031
		//
		// The NewDefaultClientConfigLoadingRules function has all the logic built
		// inside of it that handles this special case.
		clientConfLoadingrules := clientcmd.NewDefaultClientConfigLoadingRules()

		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientConfLoadingrules,
			&clientcmd.ConfigOverrides{}).ClientConfig()
	}
	if err != nil {
		return nil, err
	}

	// lower the timeout value
	restConfig.Timeout = time.Duration(timeout) * time.Second

	// create the clientset
	return kubernetes.NewForConfig(restConfig)
}
