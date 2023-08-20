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

package v1

import (
	corev1 "k8s.io/api/core/v1"
)

type StorageSpec struct {
	// Name of a persistent volume claim (PVC)
	// +optional
	Claim string `json:"claim,omitempty"`

	// Name of a storage class (to auto-provision a PV / PVC)
	// +optional
	StorageClass string `json:"storageClass,omitempty"`
}

type ConfigSpec struct {
	// Name of the secret containing the configuration for this container
	// +optional
	SecretName string `json:"secretName,omitempty"`
}

type ContainerSpec struct {
	// Container image to use to run this task
	// Default: registry.allisn.net/vpn:latest
	// +optional
	Image string `json:"image,omitempty"`

	// Restart policy if the job fails or otherwise doesn't complete
	// Default: OnFailure
	// +optional
	RestartPolicy corev1.RestartPolicy `json:"restartPolicy,omitempty"`

	// Command to execute for in this container
	// Default: ["tail", "-f", "/dev/null"]
	// +optional
	Command []string `json:"command,omitempty"`
}
