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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TaskState describes if a task has been created, is currently executing,
// finished, or failed for some other reason
// +kubebuilder:validation:Enum=Created;Scheduled;Starting;Running;Finished;Failed
type TaskState string

const (
	Created   TaskState = "Created"
	Scheduled TaskState = "Scheduled"
	Starting  TaskState = "Starting"
	Running   TaskState = "Running"
	Finished  TaskState = "Finished"
	Failed    TaskState = "Failed"
)

// TaskSpec defines the desired state of Task
type TaskSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Optional time to wait for until this task starts
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Amount of time to attempt to start this task after the
	// start time has passed
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	Deadline *metav1.Duration `json:"deadline,omitempty"`

	// Number of times the underlying job can fail before it's
	// marked failed
	// +optional
	Limit *int32 `json:"limit"`

	// Container specific items
	// +optional
	Container *ContainerSpec `json:"container,omitempty"`

	// VPN configuration details
	Config *ConfigSpec `json:"config,omitempty"`

	// What to use as the storage/filesystem area for collected data
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`
}

// TaskStatus defines the observed state of Task
type TaskStatus struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Unique id generated for this Job
	// +optional
	JobId string `json:"jobId,omitempty"`

	// Reference to the job running this task
	// +optional
	Job *corev1.ObjectReference `json:"job,omitempty"`

	// Number of attempts at this state/stage
	// +optional
	Attempt *int32 `json:"attempt,omitempty"`

	// State of this task
	State TaskState `json:"state,omitempty"`

	// Description / Reason for failed State
	// +optional
	Reason string `json:"reason,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Task is the Schema for the tasks API
type Task struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TaskSpec   `json:"spec,omitempty"`
	Status TaskStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TaskList contains a list of Task
type TaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Task `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Task{}, &TaskList{})
}
