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

// +kubebuilder:validation:Enum=Awaiting;Success;Reject;Error

const (
	NamespaceClaimStatusTypeAwaiting = "Awaiting"
	NamespaceClaimStatusTypeSuccess  = "Success"
	NamespaceClaimStatusTypeReject   = "Reject"
	NamespaceClaimStatueTypeError    = "Error"
	NamespaceClaimStatueTypeDeleted  = "Deleted"
)

// NamespaceClaimStatus defines the observed state of NamespaceClaim
type NamespaceClaimStatus struct {
	Message            string      `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	Reason             string      `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`

	// +kubebuilder:validation:Enum=Awaiting;Success;Reject;Error;
	Status string `json:"status,omitempty" protobuf:"bytes,4,opt,name=status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=nsc
// +kubebuilder:printcolumn:name="ResourceName",type=string,JSONPath=`.resourceName`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.reason`
// NamespaceClaim is the Schema for the namespaceclaims API
type NamespaceClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	ResourceName      string               `json:"resourceName"`
	Spec              v1.ResourceQuotaSpec `json:"spec,omitempty"`
	Status            NamespaceClaimStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NamespaceClaimList contains a list of NamespaceClaim
type NamespaceClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NamespaceClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NamespaceClaim{}, &NamespaceClaimList{})
}
