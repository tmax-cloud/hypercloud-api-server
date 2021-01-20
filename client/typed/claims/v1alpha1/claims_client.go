package v1alpha1

import (
	v1alpha1 "github.com/tmax-cloud/claim-operator/api/v1alpha1"

	"k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
)

type ClaimsV1alpha1Interface interface {
	RESTClient() rest.Interface
	ClusterClaimGetter
}

// ClaimsV1alpha1Client is used to interact with features provided by the  group.
type ClaimsV1alpha1Client struct {
	restClient rest.Interface
}

func (c *ClaimsV1alpha1Client) ClusterClaims(namespace string) ClusterClaimInterface {
	return newClusterClaims(c, namespace)
}

// NewForConfig creates a new ClaimsV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*ClaimsV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &ClaimsV1alpha1Client{client}, nil
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
func (c *ClaimsV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
