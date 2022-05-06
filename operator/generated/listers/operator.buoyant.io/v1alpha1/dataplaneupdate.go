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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/buoyantio/linkerd-buoyant/operator/apis/operator.buoyant.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// DataPlaneUpdateLister helps list DataPlaneUpdates.
// All objects returned here must be treated as read-only.
type DataPlaneUpdateLister interface {
	// List lists all DataPlaneUpdates in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.DataPlaneUpdate, err error)
	// DataPlaneUpdates returns an object that can list and get DataPlaneUpdates.
	DataPlaneUpdates(namespace string) DataPlaneUpdateNamespaceLister
	DataPlaneUpdateListerExpansion
}

// dataPlaneUpdateLister implements the DataPlaneUpdateLister interface.
type dataPlaneUpdateLister struct {
	indexer cache.Indexer
}

// NewDataPlaneUpdateLister returns a new DataPlaneUpdateLister.
func NewDataPlaneUpdateLister(indexer cache.Indexer) DataPlaneUpdateLister {
	return &dataPlaneUpdateLister{indexer: indexer}
}

// List lists all DataPlaneUpdates in the indexer.
func (s *dataPlaneUpdateLister) List(selector labels.Selector) (ret []*v1alpha1.DataPlaneUpdate, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.DataPlaneUpdate))
	})
	return ret, err
}

// DataPlaneUpdates returns an object that can list and get DataPlaneUpdates.
func (s *dataPlaneUpdateLister) DataPlaneUpdates(namespace string) DataPlaneUpdateNamespaceLister {
	return dataPlaneUpdateNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// DataPlaneUpdateNamespaceLister helps list and get DataPlaneUpdates.
// All objects returned here must be treated as read-only.
type DataPlaneUpdateNamespaceLister interface {
	// List lists all DataPlaneUpdates in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.DataPlaneUpdate, err error)
	// Get retrieves the DataPlaneUpdate from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.DataPlaneUpdate, error)
	DataPlaneUpdateNamespaceListerExpansion
}

// dataPlaneUpdateNamespaceLister implements the DataPlaneUpdateNamespaceLister
// interface.
type dataPlaneUpdateNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all DataPlaneUpdates in the indexer for a given namespace.
func (s dataPlaneUpdateNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.DataPlaneUpdate, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.DataPlaneUpdate))
	})
	return ret, err
}

// Get retrieves the DataPlaneUpdate from the indexer for a given namespace and name.
func (s dataPlaneUpdateNamespaceLister) Get(name string) (*v1alpha1.DataPlaneUpdate, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("dataplaneupdate"), name)
	}
	return obj.(*v1alpha1.DataPlaneUpdate), nil
}
