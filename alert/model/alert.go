package alert

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type AlertSpec struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Resource string `json:"resource"`
	Message  string `json:"message"`
}

type Alert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AlertSpec `json:"spec"`
}
