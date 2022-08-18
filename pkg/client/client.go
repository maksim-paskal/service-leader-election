/*
Copyright paskal.maksim@gmail.com
Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package client

import (
	"github.com/maksim-paskal/service-leader-election/pkg/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	Clientset  *kubernetes.Clientset
	restconfig *rest.Config
)

func Init() error {
	var err error

	if len(*config.Kubeconfig) > 0 {
		restconfig, err = clientcmd.BuildConfigFromFlags("", *config.Kubeconfig)
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
