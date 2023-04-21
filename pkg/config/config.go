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
package config

import (
	"flag"
	"os"
	"time"
)

const (
	defaultPort                     = 28086
	defaultGracefullShutdownTimeout = 5 * time.Second
)

var (
	GracefullShutdownTimeout = flag.Duration("graceful-shutdown-time", defaultGracefullShutdownTimeout, "graceful shutdown time") //nolint:lll
	LeaseName                = flag.String("lease-name", "service-leader-election", "name of lease to be created")
	CanPatchService          = flag.Bool("patch-service", true, "patch service with pod labels")
	ServiceName              = flag.String("service-name", "service-leader-election", "name of service to be patch")
	ServiceKey               = flag.String("service-key", "service-leader-election", "key of service to be patch")
	Podname                  = flag.String("podname", os.Getenv("POD_NAME"), "name of the pod")
	Namespace                = flag.String("namespace", os.Getenv("POD_NAMESPACE"), "namespace of the pod")
	ContainerName            = flag.String("container-name", "", "name of the container to watch")
	Kubeconfig               = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	Port                     = flag.Int("web.port", defaultPort, "port to listen on")
	CheckForReady            = flag.Bool("check-for-ready", true, "check for pod primary container readiness")
	CheckForReadyInterval    = flag.Duration("check-for-ready-interval", 1, "interval to check for pod primary container readiness") //nolint:lll
)
