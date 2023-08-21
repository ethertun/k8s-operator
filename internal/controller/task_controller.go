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
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	netv1 "github.com/ethertun/k8s-operator/api/v1"
)

type realClock struct{}

func (_ realClock) Now() time.Time { return time.Now() }

type Clock interface {
	Now() time.Time
}

// TaskReconciler reconciles a Task object
type TaskReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Clock
}

const (
	MAX_ATTEMPT    int32         = 5  // number of attempts to try reconciliation before marking as failed
	CHECK_INTERVAL time.Duration = 60 // time to wait in between checks when in the "running" state
)

//+kubebuilder:rbac:groups=net.ethertun.com,resources=tasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=net.ethertun.com,resources=tasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=net.ethertun.com,resources=tasks/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *TaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var task netv1.Task
	if err := r.Get(ctx, req.NamespacedName, &task); err != nil {
		if apierrors.IsNotFound(err) {
			// we can ignore not found errors, no need to requeue
			return ctrl.Result{}, nil
		}

		log.Error(err, "unable to get task")
		return ctrl.Result{}, err
	}

	// ensure the attempt value is set
	if task.Status.Attempt == nil {
		task.Status.Attempt = new(int32)
		*task.Status.Attempt = 1
	} else if *task.Status.Attempt >= MAX_ATTEMPT {
		// set status to failed
		task.Status.State = netv1.Failed
		task.Status.Reason = fmt.Sprintf("Too many attempts (%d / %d attempts)", *task.Status.Attempt, MAX_ATTEMPT)
	}

	var err error = nil
	var result ctrl.Result = ctrl.Result{Requeue: false}
	switch state := task.Status.State; state {
	case netv1.Scheduled:
		err = r.reconcileScheduled(ctx, &task)
	case netv1.Starting:
		log.V(1).Info("starting task", "jobId", task.Status.JobId)
		err = r.reconcileStarting(ctx, &task)
	case netv1.Running:
		log.V(1).Info("checking task", "jobId", task.Status.JobId)

		var requeue = false
		requeue, err = r.reconcileRunning(ctx, &task)
		if requeue {
			// assuming no errors, check again in a minute
			result = ctrl.Result{RequeueAfter: CHECK_INTERVAL * time.Second}
		}
	case netv1.Finished:
		log.V(1).Info("task finished")
	case netv1.Failed:
		log.V(0).Info("task failed", "attempts", *task.Status.Attempt, "start", task.Spec.StartTime, "deadline", task.Spec.Deadline)
	default:
		// also same as "Created"
		jobId := genJobId(5)
		log.V(1).Info("creating task", "jobId", jobId)

		task.Status.State = netv1.Scheduled
		task.Status.JobId = jobId
		result = ctrl.Result{RequeueAfter: task.Spec.StartTime.Sub(r.Now())}
	}

	if err != nil {
		log.Error(err, "unable to reconcile task")

		// bump the attempt count due a failed operation
		*task.Status.Attempt += 1
	}

	err = r.Status().Update(ctx, &task)
	if err != nil {
		log.Error(err, "unable to update task status")
	}

	return result, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *TaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// set up a real clock, since we're not in a test
	if r.Clock == nil {
		r.Clock = realClock{}
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&netv1.Task{}).
		Complete(r)
}

func (r *TaskReconciler) reconcileScheduled(ctx context.Context, task *netv1.Task) error {
	now := r.Now()
	max := task.Spec.StartTime.Add(task.Spec.Deadline.Duration)

	// check if we should run this task
	if task.Spec.StartTime.After(now) {
		// No status updates...don't run this task...yet
	} else if max.Before(now) {
		// we missed the window to run this task, mark it as failed
		start := task.Spec.StartTime.Format(time.RFC3339)
		end := max.Format(time.RFC3339)

		task.Status.State = netv1.Failed
		task.Status.Reason = fmt.Sprintf("Missed time window (start: %s, end: %s)", start, end)
	} else {
		// ok, we should run this job, update status to starting
		task.Status.State = netv1.Starting
		*task.Status.Attempt = 1
	}

	return nil
}

func (r *TaskReconciler) reconcileStarting(ctx context.Context, task *netv1.Task) error {
	name := fmt.Sprintf("%s-%s", task.Name, task.Status.JobId)
	podSpec := constructPodSpec(task.Spec.Container, task.Spec.Storage, task.Spec.Config, name)

	job := &batchv1.Job{
		ObjectMeta: construtObjectMeta(name, task.Namespace, emptyMap(), emptyMap()),
		Spec: batchv1.JobSpec{
			BackoffLimit: task.Spec.Limit,
			Template: corev1.PodTemplateSpec{
				Spec: podSpec,
			},
		},
	}

	if err := ctrl.SetControllerReference(task, job, r.Scheme); err != nil {
		return err
	}

	// attempt to create job
	if err := r.Create(ctx, job); err != nil {
		return err
	}

	jobRef, err := ref.GetReference(r.Scheme, job)
	if err != nil {
		return err
	}

	task.Status.State = netv1.Running
	task.Status.Job = jobRef
	*task.Status.Attempt = 1
	return nil
}

// Checks the current status of the job associated with the task and
// returns a tuple containing if we should requeue this check (i.e.,
// the job has not finished) or if an error occured
func (r *TaskReconciler) reconcileRunning(ctx context.Context, task *netv1.Task) (bool, error) {
	key := types.NamespacedName{
		Name:      task.Status.Job.Name,
		Namespace: task.Status.Job.Namespace,
	}

	// get the currently running job
	var job batchv1.Job
	if err := r.Get(ctx, key, &job); err != nil {
		return false, err
	}

	if job.Status.Failed > 0 {
		// find the failed condition
		var reason string = "unknown fail reason"
		for _, c := range job.Status.Conditions {
			if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
				reason = c.Message
			}
		}

		// set status as failed
		now := metav1.Now()
		task.Status.State = netv1.Failed
		task.Status.Reason = reason
		task.Status.CompletionTime = &now

		// don't return an error as the pod failing isn't a reconciliation error
		// just a regular error
		return false, nil
	} else if job.Status.CompletionTime != nil {
		task.Status.State = netv1.Finished
		task.Status.CompletionTime = job.Status.CompletionTime
		return false, nil
	}

	return true, nil
}
