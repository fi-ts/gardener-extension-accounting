package config

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	healthcheckconfig "github.com/gardener/gardener/extensions/pkg/apis/config"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControllerConfiguration defines the configuration for the fi-ts accounting provider.
type ControllerConfiguration struct {
	metav1.TypeMeta

	// Accounting is the configuration for fi-ts specific accounting in the cluster.
	Accounting Accounting

	// HealthCheckConfig is the config for the health check controller
	HealthCheckConfig *healthcheckconfig.HealthCheckConfig

	// ImagePullSecret provides an opportunity to inject an image pull secret into the resource deployments
	ImagePullSecret *ImagePullSecret
}

// Accounting contains the configuration for fi-ts specific accounting in the cluster.
type Accounting struct {
	MetalURL      string
	MetalHMAC     string
	MetalAuthType string

	// AccountingHost the host domain to reach the accounting-api
	AccountingHost string
	// AccountingPort the port to reach the accounting-api
	AccountingPort string
	// CA is the ca certificate of the accounting-api
	CA string
	// ClientCert is the client certificate to communicate with the accounting-api
	ClientCert string
	// ClientKey is the client key certificate to communicate with the accounting-api
	ClientKey string
}

// ImagePullSecret provides an opportunity to inject an image pull secret into the resource deployments
type ImagePullSecret struct {
	// DockerConfigJSON contains the already base64 encoded JSON content for the image pull secret
	DockerConfigJSON string
}
