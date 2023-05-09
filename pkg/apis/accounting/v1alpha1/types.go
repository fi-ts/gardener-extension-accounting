package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SeedAccountingResourceName  = "extension-fits-accounting"
	ShootAccountingResourceName = "extension-fits-accounting-shoot"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AccountingConfig configuration resource
type AccountingConfig struct {
	metav1.TypeMeta `json:",inline"`
}
