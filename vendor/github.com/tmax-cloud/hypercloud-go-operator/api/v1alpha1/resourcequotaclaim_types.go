/*


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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ResourceQuotaClaimStatusType defines ResourceQuotaClaim status type
// type ResourceQuotaClaimStatusType string

const (
	ResourceQuotaClaimStatusTypeAwaiting = "Awaiting"
	ResourceQuotaClaimStatusTypeSuccess  = "Success"
	ResourceQuotaClaimStatusTypeReject   = "Reject"
	ResourceQuotaClaimStatusTypeError    = "Error"
)

// ResourceQuotaClaimStatus defines the observed state of ResourceQuotaClaim
type ResourceQuotaClaimStatus struct {
	Message            string      `json:"message,omitempty" protobuf:"bytes,1,opt,name=message"`
	Reason             string      `json:"reason,omitempty" protobuf:"bytes,2,opt,name=reason"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`

	// +kubebuilder:validation:Enum=Awaiting;Success;Reject;Error;
	Status string `json:"status,omitempty" protobuf:"bytes,4,opt,name=status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rqc
// +kubebuilder:printcolumn:name="ResourceName",type=string,JSONPath=`.resourceName`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.reason`
// ResourceQuotaClaim is the Schema for the resourcequotaclaims API
type ResourceQuotaClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=.metadata"`
	ResourceName      string                   `json:"resourceName"`
	Spec              v1.ResourceQuotaSpec     `json:"spec,omitempty"`
	Status            ResourceQuotaClaimStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceQuotaClaimList contains a list of ResourceQuotaClaim
type ResourceQuotaClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceQuotaClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceQuotaClaim{}, &ResourceQuotaClaimList{})
}
