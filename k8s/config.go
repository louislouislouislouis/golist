package k8s

import (
	"flag"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/louislouislouislouis/repr8ducer/utils"
)

var (
	once       sync.Once
	k8sService *K8sService
)

func initService() {
	var err error
	k8sService, err = NewK8sService()
	if err != nil {
		utils.Log.Fatal().Stack().Err(err).Msg("Error Creating Service")
	}
}

func GetService() *K8sService {
	if k8sService == nil {
		once.Do(
			func() {
				utils.Log.WithLevel(zerolog.DebugLevel).Msg("Initializing Service")
				initService()
				utils.Log.WithLevel(zerolog.DebugLevel).Msg("Service initialized")
			},
		)
	} else {
		utils.Log.WithLevel(zerolog.DebugLevel).Msg("Sevice Already Initialized")
	}

	return k8sService
}

func NewK8sService() (*K8sService, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String(
			"kubeconfig",
			filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file",
		)
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k8s := &K8sService{
		Client: clientset,
		config: kubeconfig,
	}

	return k8s, nil
}
