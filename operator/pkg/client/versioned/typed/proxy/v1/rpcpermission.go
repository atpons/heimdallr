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
// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"time"

	v1 "github.com/f110/lagrangian-proxy/operator/pkg/api/proxy/v1"
	scheme "github.com/f110/lagrangian-proxy/operator/pkg/client/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// RpcPermissionsGetter has a method to return a RpcPermissionInterface.
// A group's client should implement this interface.
type RpcPermissionsGetter interface {
	RpcPermissions(namespace string) RpcPermissionInterface
}

// RpcPermissionInterface has methods to work with RpcPermission resources.
type RpcPermissionInterface interface {
	Create(*v1.RpcPermission) (*v1.RpcPermission, error)
	Update(*v1.RpcPermission) (*v1.RpcPermission, error)
	UpdateStatus(*v1.RpcPermission) (*v1.RpcPermission, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.RpcPermission, error)
	List(opts metav1.ListOptions) (*v1.RpcPermissionList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.RpcPermission, err error)
	RpcPermissionExpansion
}

// rpcPermissions implements RpcPermissionInterface
type rpcPermissions struct {
	client rest.Interface
	ns     string
}

// newRpcPermissions returns a RpcPermissions
func newRpcPermissions(c *ProxyV1Client, namespace string) *rpcPermissions {
	return &rpcPermissions{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the rpcPermission, and returns the corresponding rpcPermission object, and an error if there is any.
func (c *rpcPermissions) Get(name string, options metav1.GetOptions) (result *v1.RpcPermission, err error) {
	result = &v1.RpcPermission{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("rpcpermissions").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of RpcPermissions that match those selectors.
func (c *rpcPermissions) List(opts metav1.ListOptions) (result *v1.RpcPermissionList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.RpcPermissionList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("rpcpermissions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested rpcPermissions.
func (c *rpcPermissions) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("rpcpermissions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a rpcPermission and creates it.  Returns the server's representation of the rpcPermission, and an error, if there is any.
func (c *rpcPermissions) Create(rpcPermission *v1.RpcPermission) (result *v1.RpcPermission, err error) {
	result = &v1.RpcPermission{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("rpcpermissions").
		Body(rpcPermission).
		Do().
		Into(result)
	return
}

// Update takes the representation of a rpcPermission and updates it. Returns the server's representation of the rpcPermission, and an error, if there is any.
func (c *rpcPermissions) Update(rpcPermission *v1.RpcPermission) (result *v1.RpcPermission, err error) {
	result = &v1.RpcPermission{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("rpcpermissions").
		Name(rpcPermission.Name).
		Body(rpcPermission).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *rpcPermissions) UpdateStatus(rpcPermission *v1.RpcPermission) (result *v1.RpcPermission, err error) {
	result = &v1.RpcPermission{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("rpcpermissions").
		Name(rpcPermission.Name).
		SubResource("status").
		Body(rpcPermission).
		Do().
		Into(result)
	return
}

// Delete takes name of the rpcPermission and deletes it. Returns an error if one occurs.
func (c *rpcPermissions) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("rpcpermissions").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *rpcPermissions) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("rpcpermissions").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched rpcPermission.
func (c *rpcPermissions) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.RpcPermission, err error) {
	result = &v1.RpcPermission{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("rpcpermissions").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}