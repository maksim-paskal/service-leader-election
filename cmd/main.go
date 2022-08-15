package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"time"

	"github.com/maksim-paskal/service-leader-election/pkg/client"
	"github.com/maksim-paskal/service-leader-election/pkg/web"
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

var (
	podname   = flag.String("podname", os.Getenv("POD_NAME"), "name of the pod")
	namespace = flag.String("namespace", os.Getenv("POD_NAMESPACE"), "namespace of the pod")
)

var errNoPodNamespaceOrPodName = errors.New("no pod namespace or pod name")

func main() {
	flag.Parse()

	log.SetReportCaller(true)

	// loads the kubeconfig file
	if err := client.Init(); err != nil {
		log.WithError(err).Fatal("failed to init client")
	}

	// run web server for rediness and liveliness probes
	go web.StartServer()

	// run leader election
	RunLeaderElection()
}

func RunLeaderElection() {
	if len(*namespace) == 0 || len(*podname) == 0 {
		log.Fatal(errNoPodNamespaceOrPodName)
	}

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "service-leader-election",
			Namespace: *namespace,
		},
		Client: client.Clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: *podname,
		},
	}

	leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   defaultLeaseDuration,
		RenewDeadline:   defaultRenewDeadline,
		RetryPeriod:     defaultRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(c context.Context) {
				log.Info("I am the leader")
				web.IsMaster.Store(true)
			},
			OnStoppedLeading: func() {
				log.Warn("I am not the leader")
				web.IsMaster.Store(false)
			},
		},
	})
}
