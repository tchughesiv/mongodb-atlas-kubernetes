/*
Copyright 2020 MongoDB.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1/status"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/util/kube"
)

// Important:
// The procedure working with this file:
// 1. Edit the file
// 1. Run "make generate" to regenerate code
// 2. Run "make manifests" to regenerate the CRD

// Dev note: this file should be placed in "v1" package (not the nested one) as 'make manifests' doesn't generate the proper
// CRD - this may be addressed later as having a subpackage may get a much nicer code

func init() {
	SchemeBuilder.Register(&AtlasService{}, &AtlasServiceList{})
}

// AtlasServiceSpec defines the desired state of Project in Atlas
type AtlasServiceSpec struct {

	// Name is the name of the Project that is created in Atlas by the Operator if it doesn't exist yet.
	Name string `json:"name"`

	// ConnectionSecret is the name of the Kubernetes Secret which contains the information about the way to connect to
	// Atlas (organization ID, API keys). The default Operator connection configuration will be used if not provided.
	// +optional
	ConnectionSecret *ResourceRef `json:"connectionSecretRef,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:subresource:status
// +groupName:=atlas.mongodb.com

// AtlasService is the Schema for the AtlasServices API
type AtlasService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AtlasServiceSpec          `json:"spec,omitempty"`
	Status status.AtlasServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AtlasServiceList contains a list of AtlasService
type AtlasServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AtlasService `json:"items"`
}

func (p *AtlasService) ConnectionSecretObjectKey() *client.ObjectKey {
	if p.Spec.ConnectionSecret != nil {
		key := kube.ObjectKey(p.Namespace, p.Spec.ConnectionSecret.Name)
		return &key
	}
	return nil
}

func (p *AtlasService) GetStatus() status.Status {
	return p.Status
}

func (p *AtlasService) UpdateStatus(conditions []status.Condition, options ...status.Option) {
	p.Status.Conditions = conditions
	p.Status.ObservedGeneration = p.ObjectMeta.Generation

	for _, o := range options {
		// This will fail if the Option passed is incorrect - which is expected
		v := o.(status.AtlasServiceStatusOption)
		v(&p.Status)
	}
}

// ************************************ Builder methods *************************************************

func NewService(namespace, name, nameInAtlas string) *AtlasService {
	return &AtlasService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: AtlasServiceSpec{
			Name: nameInAtlas,
		},
	}
}

func (p *AtlasService) WithName(name string) *AtlasService {
	p.Name = name
	return p
}

func (p *AtlasService) WithAtlasName(name string) *AtlasService {
	p.Spec.Name = name
	return p
}

func (p *AtlasService) WithConnectionSecret(name string) *AtlasService {
	if name != "" {
		p.Spec.ConnectionSecret = &ResourceRef{Name: name}
	}
	return p
}

func DefaultService(namespace, connectionSecretName string) *AtlasService {
	return NewService(namespace, "test-service", namespace).WithConnectionSecret(connectionSecretName)
}
