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

package v1alpha2

import (
	v1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CertificateRequestLister helps list CertificateRequests.
// All objects returned here must be treated as read-only.
type CertificateRequestLister interface {
	// List lists all CertificateRequests in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.CertificateRequest, err error)
	// CertificateRequests returns an object that can list and get CertificateRequests.
	CertificateRequests(namespace string) CertificateRequestNamespaceLister
	CertificateRequestListerExpansion
}

// certificateRequestLister implements the CertificateRequestLister interface.
type certificateRequestLister struct {
	indexer cache.Indexer
}

// NewCertificateRequestLister returns a new CertificateRequestLister.
func NewCertificateRequestLister(indexer cache.Indexer) CertificateRequestLister {
	return &certificateRequestLister{indexer: indexer}
}

// List lists all CertificateRequests in the indexer.
func (s *certificateRequestLister) List(selector labels.Selector) (ret []*v1alpha2.CertificateRequest, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.CertificateRequest))
	})
	return ret, err
}

// CertificateRequests returns an object that can list and get CertificateRequests.
func (s *certificateRequestLister) CertificateRequests(namespace string) CertificateRequestNamespaceLister {
	return certificateRequestNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CertificateRequestNamespaceLister helps list and get CertificateRequests.
// All objects returned here must be treated as read-only.
type CertificateRequestNamespaceLister interface {
	// List lists all CertificateRequests in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.CertificateRequest, err error)
	// Get retrieves the CertificateRequest from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha2.CertificateRequest, error)
	CertificateRequestNamespaceListerExpansion
}

// certificateRequestNamespaceLister implements the CertificateRequestNamespaceLister
// interface.
type certificateRequestNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all CertificateRequests in the indexer for a given namespace.
func (s certificateRequestNamespaceLister) List(selector labels.Selector) (ret []*v1alpha2.CertificateRequest, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.CertificateRequest))
	})
	return ret, err
}

// Get retrieves the CertificateRequest from the indexer for a given namespace and name.
func (s certificateRequestNamespaceLister) Get(name string) (*v1alpha2.CertificateRequest, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha2.Resource("certificaterequest"), name)
	}
	return obj.(*v1alpha2.CertificateRequest), nil
}
