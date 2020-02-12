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

package v1

import (
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretSelector struct {
	Name string `json:"name"`
	Key  string `json:"key,omitempty"`
}

type LabelSelector struct {
	metav1.LabelSelector `json:",inline"`
	Namespace            string `json:"namespace,omitempty"`
}

// ProxySpec defines the desired state of Proxy
type ProxySpec struct {
	Domain   string `json:"domain"`
	Port     int32  `json:"port,omitempty"`
	HttpPort int32  `json:"httpPort,omitempty"`
	Version  string `json:"version,omitempty"`
	// Name of proxy. if not present, uses "Lagrangian Proxy CA"
	Name                  string                 `json:"name,omitempty"`
	Organization          string                 `json:"organization"`
	AdministratorUnit     string                 `json:"administratorUnit"`
	Country               string                 `json:"country,omitempty"`
	IssuerRef             cmmeta.ObjectReference `json:"issuerRef"`
	IdentityProvider      IdentityProviderSpec   `json:"identityProvider"`
	RootUsers             []string               `json:"rootUsers,omitempty"`
	Session               SessionSpec            `json:"session"`
	Replicas              int32                  `json:"replicas"`
	BackendSelector       LabelSelector          `json:"backendSelector,omitempty"`
	RoleSelector          LabelSelector          `json:"roleSelector,omitempty"`
	RpcPermissionSelector LabelSelector          `json:"rpcPermissionSelector,omitempty"`
	Defragment            DefragmentSpec         `json:"defragment,omitempty"`
	Monitor               MonitorSpec            `json:"monitor,omitempty"`
	Backup                BackupSpec             `json:"backup,omitempty"`
}

type IdentityProviderSpec struct {
	Provider        string         `json:"provider"`
	ClientId        string         `json:"clientId,omitempty"`
	ClientSecretRef SecretSelector `json:"clientSecretRef,omitempty"`
	RedirectUrl     string         `json:"redirectUrl,omitempty"`
}

type SessionSpec struct {
	Type         string         `json:"type"`
	KeySecretRef SecretSelector `json:"keySecretRef,omitempty"`
}

type DefragmentSpec struct {
	Schedule string `json:"schedule,omitempty"`
}

type MonitorSpec struct {
	// PrometheusMonitoring is set to true, then operator creates ServiceMonitor object.
	PrometheusMonitoring bool              `json:"prometheusMonitoring,omitempty"`
	Labels               map[string]string `json:"labels,omitempty"`
}

type BackupSpec struct {
	IntervalInSecond int64          `json:"intervalInSecond"`
	MaxBackups       int            `json:"maxBackups,omitempty"`
	Bucket           string         `json:"bucket"`
	Path             string         `json:"path"`
	CredentialRef    SecretSelector `json:"credentialRef"`
	Endpoint         string         `json:"endpoint,omitempty"`
}

// ProxyStatus defines the observed state of Proxy
type ProxyStatus struct {
	Ready               bool   `json:"ready"`
	Phase               string `json:"phase,omitempty"`
	NumOfBackends       int    `json:"numberOfBackends,omitempty"`
	NumOfRoles          int    `json:"numberOfRoles,omitempty"`
	NumOfRpcPermissions int    `json:"numberOfRpcPermissions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="phase",type="string",JSONPath=".status.phase",description="Phase",format="byte",priority=0
// +kubebuilder:printcolumn:name="age",type="date",JSONPath=".metadata.creationTimestamp",description="age",format="date",priority=0

// Proxy is the Schema for the proxies API
type Proxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProxySpec   `json:"spec,omitempty"`
	Status ProxyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProxyList contains a list of Proxy
type ProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Proxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Proxy{}, &ProxyList{})
}
