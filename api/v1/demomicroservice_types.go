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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DemoMicroServiceSpec defines the desired state of DemoMicroService
type DemoMicroServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Image 是该微服务容器的镜像地址，该属性不可被缺省
	Image string `json:"image"`
}

// DemoMicroServiceStatus defines the observed state of DemoMicroService
type DemoMicroServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=dms

// DemoMicroService is the Schema for the demomicroservices API
type DemoMicroService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DemoMicroServiceSpec   `json:"spec,omitempty"`
	Status DemoMicroServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DemoMicroServiceList contains a list of DemoMicroService
type DemoMicroServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DemoMicroService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DemoMicroService{}, &DemoMicroServiceList{})
}
