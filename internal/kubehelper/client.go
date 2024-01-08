package kubehelper

import (
	"os"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func createKubeClient(timeout int64) (*kubernetes.Clientset, error) {
	var cliKubeconfig string
	var cliKubecontext string

	// If there are mutiple occurrences of options supplied to kubectl, the
	// last one takes precedence. The continue statements below ensures kuberlr
	// preserves this behaviour
	for index := 1; index < len(os.Args); index++ {
		if index+1 < len(os.Args) && os.Args[index] == "--context" {
			cliKubecontext = os.Args[index+1]
			continue
		}
		if strings.HasPrefix(os.Args[index], "--context=") {
			cliKubecontext = strings.TrimPrefix(os.Args[index], "--context=")
			continue
		}
		if index+1 < len(os.Args) && os.Args[index] == "--kubeconfig" {
			cliKubeconfig = os.Args[index+1]
			continue
		}
		if strings.HasPrefix(os.Args[index], "--kubeconfig=") {
			cliKubeconfig = strings.TrimPrefix(os.Args[index], "--kubeconfig=")
			continue
		}
		if os.Args[index] == "--" {
			break
		}
	}

	// Let the NewDefaultClientConfigLoadingRules do the heavy lifting like
	// parsing the KUBECONFIG value
	// TIL: it's possible to specify multiple kubeconfig files via KUBECONFIG
	// For example: `KUBECONFIG=~/cluster1.yaml:~/cluster2.yaml`
	// See https://github.com/kubernetes/kubernetes/issues/46381#issuecomment-303926031
	//
	// The NewDefaultClientConfigLoadingRules function has all the logic built
	// inside of it that handles this special case.
	clientConfLoadingrules := clientcmd.NewDefaultClientConfigLoadingRules()
	if cliKubeconfig != "" {
		clientConfLoadingrules.ExplicitPath = cliKubeconfig
	}
	clientConfOverrides := &clientcmd.ConfigOverrides{}
	if cliKubecontext != "" {
		clientConfOverrides.CurrentContext = cliKubecontext
	}
	restConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientConfLoadingrules,
		clientConfOverrides).ClientConfig()
	if err != nil {
		return nil, err
	}

	// Lower the timeout value
	restConfig.Timeout = time.Duration(timeout) * time.Second

	// Create the clientset
	return kubernetes.NewForConfig(restConfig)
}
