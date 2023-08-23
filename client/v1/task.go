package v1

import (
	"context"
	"time"

	ethertunv1 "github.com/ethertun/k8s-operator/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type TaskInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*ethertunv1.TaskList, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*ethertunv1.Task, error)
	Create(ctx context.Context, task *ethertunv1.Task) (*ethertunv1.Task, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *ethertunv1.Task, err error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Update(ctx context.Context, task *ethertunv1.Task, opts metav1.UpdateOptions) (*ethertunv1.Task, error)
}

type taskClient struct {
	client rest.Interface
	ns     string
}

func (c *EthertunV1Client) Tasks(namespace string) TaskInterface {
	return &taskClient{
		client: c.client,
		ns:     namespace,
	}
}

// List takes label and field selectors, and returns the list of Tasks that match those selectors
func (c *taskClient) List(ctx context.Context, opts metav1.ListOptions) (*ethertunv1.TaskList, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	result := ethertunv1.TaskList{}
	err := c.client.
		Get().
		Namespace(c.ns).
		Resource("tasks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(&result)

	return &result, err
}

// Get takes the name of the task and returns the corresponding task object, and an error if there is any
func (c *taskClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*ethertunv1.Task, error) {
	result := ethertunv1.Task{}
	err := c.client.
		Get().
		Namespace(c.ns).
		Resource("tasks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

// Create takes the representation of a task and creates it.
// Returns the server's representation of the task, and an error, if there is any
func (c *taskClient) Create(ctx context.Context, task *ethertunv1.Task) (*ethertunv1.Task, error) {
	result := ethertunv1.Task{}
	err := c.client.
		Post().
		Namespace(c.ns).
		Resource("tasks").
		Body(task).
		Do(ctx).
		Into(&result)

	return &result, err
}

// Watch returns a watch.Interface that watches the requested tasks.
func (c *taskClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.
		Get().
		Namespace(c.ns).
		Resource("tasks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Update takes the representation of a task and updates it. Returns the server's representation of the task, and an error, if there is any.
func (c *taskClient) Update(ctx context.Context, task *ethertunv1.Task, opts metav1.UpdateOptions) (*ethertunv1.Task, error) {
	result := ethertunv1.Task{}
	err := c.client.
		Put().
		Namespace(c.ns).
		Resource("tasks").
		Name(task.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(task).
		Do(ctx).
		Into(&result)

	return &result, err
}

// Delete takes the name of the task and deletes it.  Returns an error if one occurs
func (c *taskClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.
		Delete().
		Namespace(c.ns).
		Resource("tasks").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of tasks.
func (c *taskClient) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}

	return c.client.
		Delete().
		Namespace(c.ns).
		Resource("tasks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched task
func (c *taskClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*ethertunv1.Task, error) {
	result := &ethertunv1.Task{}
	err := c.client.
		Patch(pt).
		Namespace(c.ns).
		Resource("tasks").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)

	return result, err
}
