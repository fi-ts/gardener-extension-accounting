//go:generate sh -c "../../vendor/github.com/gardener/gardener/hack/generate-controller-registration.sh fits-accounting . $(cat ../../VERSION) ../../example/controller-registration.yaml Extension:fits-accounting"

// Package chart enables go:generate support for generating the correct controller registration.
package chart
