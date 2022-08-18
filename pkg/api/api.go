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
package api

import (
	"context"
	"encoding/json"

	"github.com/maksim-paskal/service-leader-election/pkg/client"
	"github.com/maksim-paskal/service-leader-election/pkg/config"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// get master container ready status.
func CheckContainerIsReady(ctx context.Context) (bool, string, error) {
	pod, err := client.Clientset.CoreV1().Pods(*config.Namespace).Get(ctx, *config.Podname, metav1.GetOptions{})
	if err != nil {
		return false, "", errors.Wrap(err, "can not get pod")
	}

	if len(pod.Spec.Containers) == 1 {
		return true, pod.Spec.Containers[0].Name, nil
	}

	containerName := *config.ContainerName

	if len(containerName) == 0 {
		containerName = pod.Spec.Containers[0].Name
	}

	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == containerName {
			return container.Ready, containerName, nil
		}
	}

	return false, containerName, nil
}

func PatchService(ctx context.Context) error {
	if !*config.CanPatchService {
		return nil
	}

	service, err := client.Clientset.CoreV1().Services(*config.Namespace).Get(
		ctx,
		*config.ServiceName,
		metav1.GetOptions{},
	)
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

	service.Spec.Selector[*config.ServiceKey] = *config.Podname

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "error marshaling payload")
	}

	_, err = client.Clientset.CoreV1().Services(*config.Namespace).Patch(
		ctx,
		*config.ServiceName,
		types.StrategicMergePatchType,
		payloadBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		return errors.Wrap(err, "error patching service")
	}

	return nil
}

func PatchPodLabels(ctx context.Context) error {
	pod, err := client.Clientset.CoreV1().Pods(*config.Namespace).Get(ctx, *config.Podname, metav1.GetOptions{})
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

	pod.Labels[*config.ServiceKey] = *config.Podname

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "error marshaling payload")
	}

	_, err = client.Clientset.CoreV1().Pods(*config.Namespace).Patch(
		ctx,
		*config.Podname,
		types.StrategicMergePatchType,
		payloadBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		return errors.Wrap(err, "error patching pod")
	}

	return nil
}
