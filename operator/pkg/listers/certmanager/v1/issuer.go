/*

MIT License

Copyright (c) 2019 Fumihiro Ito

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/
// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// IssuerLister helps list Issuers.
// All objects returned here must be treated as read-only.
type IssuerLister interface {
	// List lists all Issuers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.Issuer, err error)
	// Issuers returns an object that can list and get Issuers.
	Issuers(namespace string) IssuerNamespaceLister
	IssuerListerExpansion
}

// issuerLister implements the IssuerLister interface.
type issuerLister struct {
	indexer cache.Indexer
}

// NewIssuerLister returns a new IssuerLister.
func NewIssuerLister(indexer cache.Indexer) IssuerLister {
	return &issuerLister{indexer: indexer}
}

// List lists all Issuers in the indexer.
func (s *issuerLister) List(selector labels.Selector) (ret []*v1.Issuer, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Issuer))
	})
	return ret, err
}

// Issuers returns an object that can list and get Issuers.
func (s *issuerLister) Issuers(namespace string) IssuerNamespaceLister {
	return issuerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// IssuerNamespaceLister helps list and get Issuers.
// All objects returned here must be treated as read-only.
type IssuerNamespaceLister interface {
	// List lists all Issuers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.Issuer, err error)
	// Get retrieves the Issuer from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.Issuer, error)
	IssuerNamespaceListerExpansion
}

// issuerNamespaceLister implements the IssuerNamespaceLister
// interface.
type issuerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Issuers in the indexer for a given namespace.
func (s issuerNamespaceLister) List(selector labels.Selector) (ret []*v1.Issuer, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Issuer))
	})
	return ret, err
}

// Get retrieves the Issuer from the indexer for a given namespace and name.
func (s issuerNamespaceLister) Get(name string) (*v1.Issuer, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("issuer"), name)
	}
	return obj.(*v1.Issuer), nil
}
