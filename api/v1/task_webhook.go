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
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	validationutils "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	DEFAULT_DEADLINE time.Duration = 60 * time.Second // default deadline is 1 minute (60 seconds)
	DEFAULT_LIMIT                  = 3                // default number of attempts is 3
)

// log is for logging in this package.
var tasklog = logf.Log.WithName("task-resource")

func (r *Task) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-net-ethertun-com-v1-task,mutating=true,failurePolicy=fail,sideEffects=None,groups=net.ethertun.com,resources=tasks,verbs=create;update,versions=v1,name=mtask.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Task{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Task) Default() {
	tasklog.Info("default", "name", r.Name)

	if r.Spec.Container == nil {
		r.Spec.Container = &ContainerSpec{}
	}

	if r.Spec.Container.Image == "" {
		r.Spec.Container.Image = DEFAULT_IMAGE
	}

	if r.Spec.Container.RestartPolicy == "" {
		r.Spec.Container.RestartPolicy = DEFAULT_RESTART_POLICY
	}

	if r.Spec.Config == nil {
		r.Spec.Config = &ConfigSpec{}
	}

	if r.Spec.Config.SecretName == "" {
		r.Spec.Config.SecretName = fmt.Sprintf("%s-cfg", r.Name)
	}

	if r.Spec.StartTime == nil {
		r.Spec.StartTime = new(metav1.Time)
		*r.Spec.StartTime = metav1.Now()
	}

	if r.Spec.Deadline == nil {
		r.Spec.Deadline = new(metav1.Duration)
		r.Spec.Deadline.Duration = DEFAULT_DEADLINE
	}

	if r.Spec.Limit == nil {
		r.Spec.Limit = new(int32)
		*r.Spec.Limit = DEFAULT_LIMIT
	}

}

// NOTE: change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-net-ethertun-com-v1-task,mutating=false,failurePolicy=fail,sideEffects=None,groups=net.ethertun.com,resources=tasks,verbs=create;update,versions=v1,name=vtask.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Task{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Task) ValidateCreate() (admission.Warnings, error) {
	tasklog.Info("validate create", "name", r.Name)

	return nil, r.ValidateTask()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Task) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	tasklog.Info("validate update", "name", r.Name)

	return nil, r.ValidateTask()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Task) ValidateDelete() (admission.Warnings, error) {
	tasklog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *Task) ValidateTask() error {
	var allErrs field.ErrorList

	if len(r.ObjectMeta.Name) > validationutils.DNS1035LabelMaxLength-6 {
		// ensure name has at least six spare characters to account for
		// the `-jobid` added to each task.
		err := field.Invalid(field.NewPath("metadata").Child("name"), r.Name, "must be no more than 57 characters")
		allErrs = append(allErrs, err)
	}

	if len(r.Spec.Container.Command) == 0 {
		// a command must be specified
		err := field.Required(field.NewPath("spec").Child("container").Child("command"), "a command must be specified")
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "net.ethertun.com", Kind: "Task"},
		r.Name, allErrs,
	)
}
