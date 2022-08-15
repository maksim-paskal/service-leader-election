package client

import (
	"flag"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")

var (
	Clientset  *kubernetes.Clientset
	restconfig *rest.Config
)

func Init() error {
	var err error

	if len(*kubeconfig) > 0 {
		restconfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return errors.Wrap(err, "failed to build rest config")
		}
	} else {
		log.Info("No kubeconfig file use incluster")
		restconfig, err = rest.InClusterConfig()
		if err != nil {
			return errors.Wrap(err, "failed to create cluster config")
		}
	}

	Clientset, err = kubernetes.NewForConfig(restconfig)
	if err != nil {
		log.WithError(err).Fatal()
	}

	return nil
}
