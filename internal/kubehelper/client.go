package kubehelper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flavio/kuberlr/internal/common"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func createKubeClient() (*kubernetes.Clientset, error) {
	kubeconfig := getKubeconfig()
	if kubeconfig == "" {
		return nil, fmt.Errorf(
			"Canot find kubeconfig file. Either create %s or set the KUBECONFIG environment variable\n",
			defaultKubeconfig())
	}

	// use the current context in kubeconfig
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	return kubernetes.NewForConfig(config)
}

func getKubeconfig() string {
	if k := os.Getenv("KUBECONFIG"); k != "" {
		return k
	}

	return defaultKubeconfig()
}

func defaultKubeconfig() string {
	if home := common.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}

	return ""
}
