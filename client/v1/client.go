package v1

import (
	"context"

	ethertunv1 "github.com/ethertun/k8s-operator/api/v1"

	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type EthertunV1Interface interface {
	Tasks(ctx context.Context, namespace string) TaskInterface
}

type EthertunV1Client struct {
	client rest.Interface
}

func NewForConfig(config *rest.Config) (*EthertunV1Client, error) {
	ethertunv1.AddToScheme(scheme.Scheme)
	config.ContentConfig.GroupVersion = &ethertunv1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.UnversionedRESTClientFor(config)
	if err != nil {
		return nil, err
	}

	return &EthertunV1Client{client: client}, nil
}
