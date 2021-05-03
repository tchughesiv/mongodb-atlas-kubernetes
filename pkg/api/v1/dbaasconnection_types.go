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

	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1/status"
)

// Important:
// The procedure working with this file:
// 1. Edit the file
// 1. Run "make generate" to regenerate code
// 2. Run "make manifests" to regenerate the CRD

// Dev note: this file should be placed in "v1" package (not the nested one) as 'make manifests' doesn't generate the proper
// CRD - this may be addressed later as having a subpackage may get a much nicer code

func init() {
	SchemeBuilder.Register(&DbaaSConnection{}, &DbaaSConnectionList{})
}

// DbaaSConnectionSpec defines the desired state of DbaaSConnection
type DbaaSConnectionSpec struct {

	// Name is the name of the database cluster reprented by DbaaSCluster.
	Cluster string `json:"cluster"`

	// Vendor is vendor of the DB, such as mongodb
	Vendor string `json:"vendor,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:subresource:status
// +groupName:=atlas.mongodb.com

// DbaaSConnection is the Schema for the DbaaSConnections API
type DbaaSConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DbaaSConnectionSpec          `json:"spec,omitempty"`
	Status status.DbaaSConnectionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DbaaSConnectionList contains a list of DbaaSConnection
type DbaaSConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DbaaSConnection `json:"items"`
}

func (p *DbaaSConnection) GetStatus() status.Status {
	return p.Status
}

func (p *DbaaSConnection) UpdateStatus(conditions []status.Condition, options ...status.Option) {
	p.Status.Conditions = conditions
	p.Status.ObservedGeneration = p.ObjectMeta.Generation

	for _, o := range options {
		// This will fail if the Option passed is incorrect - which is expected
		v := o.(status.DbaaSConnectionStatusOption)
		v(&p.Status)
	}
}

// ************************************ Builder methods *************************************************

func NewConnection(namespace, name, cluster string) *DbaaSConnection {
	return &DbaaSConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: DbaaSConnectionSpec{
			Cluster: cluster,
		},
	}
}
