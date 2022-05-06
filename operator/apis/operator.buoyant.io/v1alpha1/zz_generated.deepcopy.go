//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Components) DeepCopyInto(out *Components) {
	*out = *in
	if in.Linkerd != nil {
		in, out := &in.Linkerd, &out.Linkerd
		*out = new(HelmReleaseParams)
		**out = **in
	}
	if in.LinkerdViz != nil {
		in, out := &in.LinkerdViz, &out.LinkerdViz
		*out = new(HelmReleaseParams)
		**out = **in
	}
	if in.LinkerdMulticluster != nil {
		in, out := &in.LinkerdMulticluster, &out.LinkerdMulticluster
		*out = new(HelmReleaseParams)
		**out = **in
	}
	if in.LinkerdJaeger != nil {
		in, out := &in.LinkerdJaeger, &out.LinkerdJaeger
		*out = new(HelmReleaseParams)
		**out = **in
	}
	if in.LinkerdSmi != nil {
		in, out := &in.LinkerdSmi, &out.LinkerdSmi
		*out = new(HelmReleaseParams)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Components.
func (in *Components) DeepCopy() *Components {
	if in == nil {
		return nil
	}
	out := new(Components)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneUpdate) DeepCopyInto(out *ControlPlaneUpdate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneUpdate.
func (in *ControlPlaneUpdate) DeepCopy() *ControlPlaneUpdate {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneUpdate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ControlPlaneUpdate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneUpdateList) DeepCopyInto(out *ControlPlaneUpdateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ControlPlaneUpdate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneUpdateList.
func (in *ControlPlaneUpdateList) DeepCopy() *ControlPlaneUpdateList {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneUpdateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ControlPlaneUpdateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneUpdateSpec) DeepCopyInto(out *ControlPlaneUpdateSpec) {
	*out = *in
	if in.Components != nil {
		in, out := &in.Components, &out.Components
		*out = new(Components)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneUpdateSpec.
func (in *ControlPlaneUpdateSpec) DeepCopy() *ControlPlaneUpdateSpec {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneUpdateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneUpdateStatus) DeepCopyInto(out *ControlPlaneUpdateStatus) {
	*out = *in
	in.LastUpdateAttempt.DeepCopyInto(&out.LastUpdateAttempt)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneUpdateStatus.
func (in *ControlPlaneUpdateStatus) DeepCopy() *ControlPlaneUpdateStatus {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneUpdateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneUpdate) DeepCopyInto(out *DataPlaneUpdate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneUpdate.
func (in *DataPlaneUpdate) DeepCopy() *DataPlaneUpdate {
	if in == nil {
		return nil
	}
	out := new(DataPlaneUpdate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DataPlaneUpdate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneUpdateList) DeepCopyInto(out *DataPlaneUpdateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DataPlaneUpdate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneUpdateList.
func (in *DataPlaneUpdateList) DeepCopy() *DataPlaneUpdateList {
	if in == nil {
		return nil
	}
	out := new(DataPlaneUpdateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DataPlaneUpdateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneUpdateSpec) DeepCopyInto(out *DataPlaneUpdateSpec) {
	*out = *in
	if in.WorkloadSelector != nil {
		in, out := &in.WorkloadSelector, &out.WorkloadSelector
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneUpdateSpec.
func (in *DataPlaneUpdateSpec) DeepCopy() *DataPlaneUpdateSpec {
	if in == nil {
		return nil
	}
	out := new(DataPlaneUpdateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneUpdateStatus) DeepCopyInto(out *DataPlaneUpdateStatus) {
	*out = *in
	in.LastUpdateAttempt.DeepCopyInto(&out.LastUpdateAttempt)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneUpdateStatus.
func (in *DataPlaneUpdateStatus) DeepCopy() *DataPlaneUpdateStatus {
	if in == nil {
		return nil
	}
	out := new(DataPlaneUpdateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HelmReleaseParams) DeepCopyInto(out *HelmReleaseParams) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HelmReleaseParams.
func (in *HelmReleaseParams) DeepCopy() *HelmReleaseParams {
	if in == nil {
		return nil
	}
	out := new(HelmReleaseParams)
	in.DeepCopyInto(out)
	return out
}
