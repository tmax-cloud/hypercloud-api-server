package alert

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *Alert) DeepCopyInto(out *Alert) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = AlertSpec{
		Kind:     in.Spec.Kind,
		Name:     in.Spec.Name,
		Resource: in.Spec.Resource,
		Message:  in.Spec.Message,
	}
}

// DeepCopyObject returns a generically typed copy of an object
func (in *Alert) DeepCopyObject() runtime.Object {
	out := Alert{}
	in.DeepCopyInto(&out)

	return &out
}
