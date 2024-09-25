/*
Copyright 2024.

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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DomainEgressPolicySpec defines the desired state of DomainEgressPolicy
type DomainEgressPolicySpec struct {
	// podSelector selects the pods to which this DomainEgressPolicy object applies.
	// The array of ingress rules is applied to any pods selected by this field.
	// Multiple network policies can select the same set of pods. In this case,
	// the ingress rules for each are combined additively.
	// This field is NOT optional and follows standard label selector semantics.
	// An empty podSelector matches all pods in this namespace.
	PodSelector metav1.LabelSelector `json:"podSelector" protobuf:"bytes,1,opt,name=podSelector"`

	// domainEgress is a list of egress rules to be applied to the selected pods. Outgoing traffic
	// is allowed if there are no DomainEgressPolicies selecting the pod (and cluster policy
	// otherwise allows the traffic), OR if the traffic matches at least one egress rule
	// across all of the DomainEgressPolicy objects whose podSelector matches the pod. If
	// this field is empty then this DomainEgressPolicy limits all outgoing traffic (and serves
	// solely to ensure that the pods it selects are isolated by default).
	DomainEgress []DomainEgressRule `json:"domainEgress,omitempty" protobuf:"bytes,3,rep,name=domainEgress"`
}

// DomainEgressRule describes a particular set of traffic that is allowed out of pods
// matched by a DomainEgressPolicySpec's podSelector. The traffic must match both ports and domains.
type DomainEgressRule struct {
	// ports is a list of destination ports for outgoing traffic.
	// Each item in this list is combined using a logical OR. If this field is
	// empty or missing, this rule matches all ports (traffic not restricted by port).
	// If this field is present and contains at least one item, then this rule allows
	// traffic only if the traffic matches at least one port in the list.
	// +optional
	Ports []networkingv1.NetworkPolicyPort `json:"ports,omitempty" protobuf:"bytes,1,rep,name=ports"`

	// domains is a list of destinations for outgoing traffic of pods selected for this rule.
	// Items in this list are combined using a logical OR operation. If this field is
	// empty or missing, this rule matches all destinations (traffic not restricted by
	// destination). If this field is present and contains at least one item, this rule
	// allows traffic only if the traffic matches at least one item in the to list.
	// domain names may be fully qualified or match wildcards.
	Domains []string `json:"domains,omitempty" protobuf:"bytes,2,rep,name=domains"`
}

// DomainEgressPolicyStatus defines the observed state of DomainEgressPolicy
type DomainEgressPolicyStatus struct {
	// ResolvedDomains is a list of resolved domains
	ResolvedDomains []ResolvedDomain `json:"resolvedDomains,omitempty" protobuf:"bytes,1,rep,name=resolvedDomains"`
}

// ResolvedDomain defines the domain and resolved IP addresses
type ResolvedDomain struct {
	Domain string   `json:"domain" protobuf:"bytes,1,opt,name=domain"`
	IPs    []string `json:"ips" protobuf:"bytes,2,rep,name=ips"`

	// UpdateTimestamp is a timestamp of when the domain was resolved
	UpdateTimestamp metav1.Time `json:"updateTimestamp,omitempty" protobuf:"bytes,8,opt,name=updateTimestamp"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DomainEgressPolicy is the Schema for the domainegresspolicies API
type DomainEgressPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DomainEgressPolicySpec   `json:"spec,omitempty"`
	Status DomainEgressPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DomainEgressPolicyList contains a list of DomainEgressPolicy
type DomainEgressPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DomainEgressPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DomainEgressPolicy{}, &DomainEgressPolicyList{})
}
