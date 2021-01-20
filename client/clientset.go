package client

import (
	"fmt"

	claimsv1alpha1 "github.com/tmax-cloud/hypercloud-multi-api-server/client/typed/claims/v1alpha1"
	clusterv1alpha1 "github.com/tmax-cloud/hypercloud-multi-api-server/client/typed/cluster/v1alpha1"
	configv1alpha1 "github.com/tmax-cloud/hypercloud-multi-api-server/client/typed/config/v1alpha1"

	discovery "k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	ConfigV1alpha1() configv1alpha1.ConfigV1alpha1Interface
	ClaimsV1alpha1() claimsv1alpha1.ClaimsV1alpha1Interface
	ClusterV1alpha1() clusterv1alpha1.ClusterV1alpha1Interface
}

type Clientset struct {
	*discovery.DiscoveryClient
	configV1alpha1  *configv1alpha1.ConfigV1alpha1Client
	claimsv1alpha1  *claimsv1alpha1.ClaimsV1alpha1Client
	clusterv1alpha1 *clusterv1alpha1.ClusterV1alpha1Client
}

// CoreV1 retrieves the CoreV1Client
func (c *Clientset) ConfigV1alpha1() configv1alpha1.ConfigV1alpha1Interface {
	return c.configV1alpha1
}

// CoreV1 retrieves the CoreV1Client
func (c *Clientset) ClaimsV1alpha1() claimsv1alpha1.ClaimsV1alpha1Interface {
	return c.claimsv1alpha1
}

// CoreV1 retrieves the CoreV1Client
func (c *Clientset) ClusterV1alpha1() clusterv1alpha1.ClusterV1alpha1Interface {
	return c.clusterv1alpha1
}

func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error

	cs.configV1alpha1, err = configv1alpha1.NewForConfig(&configShallowCopy)
	cs.claimsv1alpha1, err = claimsv1alpha1.NewForConfig(&configShallowCopy)
	cs.clusterv1alpha1, err = clusterv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}
