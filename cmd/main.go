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
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maksim-paskal/service-leader-election/pkg/api"
	"github.com/maksim-paskal/service-leader-election/pkg/client"
	"github.com/maksim-paskal/service-leader-election/pkg/config"
	"github.com/maksim-paskal/service-leader-election/pkg/web"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

const (
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
)

var errNoPodNamespaceOrPodName = errors.New("no pod namespace or pod name")

func main() {
	// parse cli flags
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// listen for termination
	go func() {
		<-sigs
		cancel()
	}()

	log.RegisterExitHandler(func() {
		cancel()
		time.Sleep(*config.GracefullShutdownTimeout)
	})

	// log file name
	log.SetReportCaller(true)

	// check if pod namespace and pod name are set
	if len(*config.Namespace) == 0 || len(*config.Podname) == 0 {
		log.Fatal(errNoPodNamespaceOrPodName)
	}

	// loads the kubeconfig file
	if err := client.Init(); err != nil {
		log.WithError(err).Fatal("failed to init client")
	}

	// patch pod with label
	if err := api.PatchPodLabels(ctx); err != nil {
		log.WithError(err).Fatal("failed to patch pod with labels")
	}

	// wait for pod readiness
	if err := waitForPodReady(ctx); err != nil {
		log.WithError(err).Fatal("failed to wait for pod ready")
	}

	// run web server for rediness and liveliness probes
	go web.StartServer(ctx)

	// run leader election
	runLeaderElection(ctx)

	<-ctx.Done()

	time.Sleep(*config.GracefullShutdownTimeout)
}

func waitForPodReady(ctx context.Context) error {
	if !*config.CheckForReady {
		return nil
	}

	for {
		ready, container, err := api.CheckContainerIsReady(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to check container readiness")
		}

		if ready {
			return nil
		}

		log.Infof("wait for container %s will be ready", container)
		time.Sleep(*config.CheckForReadyInterval)
	}
}

func runLeaderElection(ctx context.Context) {
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      *config.LeaseName,
			Namespace: *config.Namespace,
		},
		Client: client.Clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: *config.Podname,
		},
	}

	go leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   defaultLeaseDuration,
		RenewDeadline:   defaultRenewDeadline,
		RetryPeriod:     defaultRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				log.Info("I am the leader")
				if err := api.PatchService(ctx); err != nil {
					log.WithError(err).Fatal("failed to patch service")
				}

				web.IsMaster.Store(true)
			},
			OnStoppedLeading: func() {
				log.Warn("I am not the leader")
				web.IsMaster.Store(false)
			},
		},
	})
}
