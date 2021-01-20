package v1alpha1

import (
	v1alpha1 "github.com/tmax-cloud/efk-operator/api/v1alpha1"

	"k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
)

type ConfigV1alpha1Interface interface {
	RESTClient() rest.Interface
	FluentBitConfigurationGetter
}

// ConfigV1alpha1Client is used to interact with features provided by the  group.
type ConfigV1alpha1Client struct {
	restClient rest.Interface
}

func (c *ConfigV1alpha1Client) FluentBitConfigurations(namespace string) FluentBitConfigurationInterface {
	return newFluentBitConfigurations(c, namespace)
}

// NewForConfig creates a new ConfigV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*ConfigV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &ConfigV1alpha1Client{client}, nil
}

func setConfigDefaults(config *rest.Config) error {
	v1alpha1.AddToScheme(scheme.Scheme)
	gv := v1alpha1.GroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *ConfigV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
