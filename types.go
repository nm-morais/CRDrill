package main

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceRef struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

type Spec struct {
	ResourceRefs []ResourceRef `json:"resourceRefs"`
}

type Condition struct {
	LastTransitionTime string `json:"lastTransitionTime"`
	Reason             string `json:"reason"`
	Status             string `json:"status"`
	ConditionType      string `json:"type"`
}

type Status struct {
	Conditions []Condition `json:"conditions"`
}

type CrossplaneCRD struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Status            Status `json:"status"`
	Spec              Spec   `json:"spec"`
}

func (crd CrossplaneCRD) IsReady() (bool, error) {
	hasReadyStatus := false
	for _, condition := range crd.Status.Conditions {
		if condition.ConditionType == "Ready" {
			hasReadyStatus = true
			if condition.Status == "True" {
				return true, nil
			}
		}
	}
	if hasReadyStatus {
		return false, nil
	}
	return false, NoStatusConditionErr
}

func (crd CrossplaneCRD) String() string {
	return fmt.Sprintf("Name: %s, Kind: %s, ApiVersion %s", crd.Name, crd.Kind, crd.APIVersion)
}

type CrossplaneCRDList struct {
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []CrossplaneCRD `json:"items" protobuf:"bytes,2,rep,name=items"`
}
