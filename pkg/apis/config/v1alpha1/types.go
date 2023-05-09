package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	healthcheckconfigv1alpha1 "github.com/gardener/gardener/extensions/pkg/controller/healthcheck/config/v1alpha1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControllerConfiguration defines the configuration for the fi-ts accounting provider.
type ControllerConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// Accounting is the configuration for fi-ts specific accounting in the cluster.
	Accounting Accounting `json:"accounting"`

	// HealthCheckConfig is the config for the health check controller
	// +optional
	HealthCheckConfig *healthcheckconfigv1alpha1.HealthCheckConfig `json:"healthCheckConfig,omitempty"`

	// ImagePullSecret provides an opportunity to inject an image pull secret into the resource deployments
	ImagePullSecret *ImagePullSecret `json:"imagePullSecret,omitempty"`
}

// Accounting contains the configuration for fi-ts specific accounting in the cluster.
type Accounting struct {
	MetalURL      string `json:"metalURL"`
	MetalHMAC     string `json:"metalHMAC"`
	MetalAuthType string `json:"metalAuthType"`

	CA         string `json:"ca"`
	ClientCert string `json:"cert"`
	ClientKey  string `json:"key"`
}

// ImagePullSecret provides an opportunity to inject an image pull secret into the resource deployments
type ImagePullSecret struct {
	// DockerConfigJSON contains the already base64 encoded JSON content for the image pull secret
	DockerConfigJSON string `json:"encodedDockerConfigJSON"`
}
