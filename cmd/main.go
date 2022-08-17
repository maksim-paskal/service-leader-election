package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/maksim-paskal/service-leader-election/pkg/client"
	"github.com/maksim-paskal/service-leader-election/pkg/web"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

const (
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
)

var (
	leaseName       = flag.String("lease-name", "service-leader-election", "name of lease to be created")
	canPatchService = flag.Bool("patch-service", true, "patch service with pod labels")
	serviceName     = flag.String("service-name", "service-leader-election", "name of service to be patch")
	serviceKey      = flag.String("service-key", "service-leader-election", "key of service to be patch")
	podname         = flag.String("podname", os.Getenv("POD_NAME"), "name of the pod")
	namespace       = flag.String("namespace", os.Getenv("POD_NAMESPACE"), "namespace of the pod")
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
	RunLeaderElection(context.Background())
}

func patchService(ctx context.Context) error {
	if !*canPatchService {
		return nil
	}

	service, err := client.Clientset.CoreV1().Services(*namespace).Get(ctx, *serviceName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "can not get service")
	}

	type metadataStringValue struct {
		Selector map[string]string `json:"selector"`
	}

	type patchStringValue struct {
		Spec metadataStringValue `json:"spec"`
	}

	payload := patchStringValue{
		Spec: metadataStringValue{
			Selector: service.Spec.Selector,
		},
	}

	if service.Spec.Selector == nil {
		service.Spec.Selector = map[string]string{}
	}

	service.Spec.Selector[*serviceKey] = *podname

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "error marshaling payload")
	}

	_, err = client.Clientset.CoreV1().Services(*namespace).Patch(
		ctx,
		*serviceName,
		types.StrategicMergePatchType,
		payloadBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		return errors.Wrap(err, "error patching service")
	}

	return nil
}

func patchPodLabels(ctx context.Context) error {
	pod, err := client.Clientset.CoreV1().Pods(*namespace).Get(ctx, *podname, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "can not get pod")
	}

	type metadataStringValue struct {
		Annotations map[string]string `json:"annotations"`
		Labels      map[string]string `json:"labels"`
	}

	type patchStringValue struct {
		Metadata metadataStringValue `json:"metadata"`
	}

	payload := patchStringValue{
		Metadata: metadataStringValue{
			Annotations: pod.Annotations,
			Labels:      pod.Labels,
		},
	}

	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}

	pod.Labels[*serviceKey] = *podname

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "error marshaling payload")
	}

	_, err = client.Clientset.CoreV1().Pods(*namespace).Patch(
		ctx,
		*podname,
		types.StrategicMergePatchType,
		payloadBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		return errors.Wrap(err, "error patching pod")
	}

	return nil
}

func RunLeaderElection(ctx context.Context) {
	if len(*namespace) == 0 || len(*podname) == 0 {
		log.Fatal(errNoPodNamespaceOrPodName)
	}

	if err := patchPodLabels(ctx); err != nil {
		log.WithError(err).Fatal("failed to patch pod with labels")
	}

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      *leaseName,
			Namespace: *namespace,
		},
		Client: client.Clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: *podname,
		},
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   defaultLeaseDuration,
		RenewDeadline:   defaultRenewDeadline,
		RetryPeriod:     defaultRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				log.Info("I am the leader")
				if err := patchService(ctx); err != nil {
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
