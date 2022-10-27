package main

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	NoStatusConditionErr = errors.New("CRD does not have a status condition")
)

type ResourceRef struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

type Spec struct {
	ResourceRefs []ResourceRef `json:"resourceRefs"`
}

type Status struct {
	Conditions []metav1.Condition `json:"conditions"`
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
		if condition.Type == "Ready" {
			hasReadyStatus = true
			if condition.Status == "True" {
				return true, nil
			}
		}
	}

	if hasErr, reconcileErr := crd.HasReconcileError(); hasErr {
		return false, reconcileErr
	}

	if !hasReadyStatus {
		return false, NoStatusConditionErr
	}
	return false, nil
}

func (crd CrossplaneCRD) HasReconcileError() (bool, error) {
	for _, condition := range crd.Status.Conditions {
		if condition.Type == "Synced" && condition.Reason == "ReconcileError" {
			return true, fmt.Errorf("%s", condition.Message)
		}
	}
	return false, nil
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
