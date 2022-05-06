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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	operatorbuoyantiov1alpha1 "github.com/buoyantio/linkerd-buoyant/operator/apis/operator.buoyant.io/v1alpha1"
	versioned "github.com/buoyantio/linkerd-buoyant/operator/generated/clientset/versioned"
	internalinterfaces "github.com/buoyantio/linkerd-buoyant/operator/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/buoyantio/linkerd-buoyant/operator/generated/listers/operator.buoyant.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ControlPlaneUpdateInformer provides access to a shared informer and lister for
// ControlPlaneUpdates.
type ControlPlaneUpdateInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ControlPlaneUpdateLister
}

type controlPlaneUpdateInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewControlPlaneUpdateInformer constructs a new informer for ControlPlaneUpdate type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewControlPlaneUpdateInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredControlPlaneUpdateInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredControlPlaneUpdateInformer constructs a new informer for ControlPlaneUpdate type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredControlPlaneUpdateInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.OperatorV1alpha1().ControlPlaneUpdates().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.OperatorV1alpha1().ControlPlaneUpdates().Watch(context.TODO(), options)
			},
		},
		&operatorbuoyantiov1alpha1.ControlPlaneUpdate{},
		resyncPeriod,
		indexers,
	)
}

func (f *controlPlaneUpdateInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredControlPlaneUpdateInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *controlPlaneUpdateInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&operatorbuoyantiov1alpha1.ControlPlaneUpdate{}, f.defaultInformer)
}

func (f *controlPlaneUpdateInformer) Lister() v1alpha1.ControlPlaneUpdateLister {
	return v1alpha1.NewControlPlaneUpdateLister(f.Informer().GetIndexer())
}
