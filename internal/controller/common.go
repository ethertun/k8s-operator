/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"math/rand"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	opsv1 "github.com/ethertun/k8s-operator/api/v1"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

// Generates a unique DNS-compliant identifier that can be attached
// to a container/pod/service/etc.'s name
func genJobId(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}

	return string(b)
}

func emptyMap() map[string]string {
	return make(map[string]string)
}

// Helper function to construct the meta attributes for an object
func construtObjectMeta(name, namespace string, labels map[string]string, annotations map[string]string) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Labels:      labels,
		Annotations: annotations,
		Name:        name,
		Namespace:   namespace,
	}

	return meta
}

// Helper function to build the pod spec for various kinds
func constructPodSpec(container *opsv1.ContainerSpec, storage *opsv1.StorageSpec, cfg *opsv1.ConfigSpec, name string) corev1.PodSpec {
	privileged := true

	spec := corev1.PodSpec{
		RestartPolicy: container.RestartPolicy,
		Containers: []corev1.Container{
			{
				Name:  fmt.Sprintf("%s-pod", name),
				Image: container.Image,
				Args:  container.Command,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
					Capabilities: &corev1.Capabilities{
						Add: []corev1.Capability{
							corev1.Capability("NET_ADMIN"),
						},
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "config-mount",
						MountPath: "/data",
					},
					{
						Name:      "data-mount",
						MountPath: "/data/out",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "config-mount",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: cfg.SecretName,
					},
				},
			},
			{
				Name: "data-mount",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		},
	}

	return spec
}

func constructServiceSpec(port int32, selectorId, selectorValue string) corev1.ServiceSpec {
	service := corev1.ServiceSpec{
		Type: corev1.ServiceTypeNodePort,
		Selector: map[string]string{
			selectorId: selectorValue,
		},
		Ports: []corev1.ServicePort{
			{
				Protocol:   corev1.ProtocolTCP,
				Port:       port,
				TargetPort: intstr.FromInt(int(port)),
			},
		},
	}

	return service
}
