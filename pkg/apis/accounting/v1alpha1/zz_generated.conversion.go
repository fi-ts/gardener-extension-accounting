//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
2024 Copyright FI-TS Finanz Informatik Technologie Service.
*/

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	accounting "github.com/fi-ts/gardener-extension-accounting/pkg/apis/accounting"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*AccountingConfig)(nil), (*accounting.AccountingConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_AccountingConfig_To_accounting_AccountingConfig(a.(*AccountingConfig), b.(*accounting.AccountingConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*accounting.AccountingConfig)(nil), (*AccountingConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_accounting_AccountingConfig_To_v1alpha1_AccountingConfig(a.(*accounting.AccountingConfig), b.(*AccountingConfig), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_AccountingConfig_To_accounting_AccountingConfig(in *AccountingConfig, out *accounting.AccountingConfig, s conversion.Scope) error {
	return nil
}

// Convert_v1alpha1_AccountingConfig_To_accounting_AccountingConfig is an autogenerated conversion function.
func Convert_v1alpha1_AccountingConfig_To_accounting_AccountingConfig(in *AccountingConfig, out *accounting.AccountingConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_AccountingConfig_To_accounting_AccountingConfig(in, out, s)
}

func autoConvert_accounting_AccountingConfig_To_v1alpha1_AccountingConfig(in *accounting.AccountingConfig, out *AccountingConfig, s conversion.Scope) error {
	return nil
}

// Convert_accounting_AccountingConfig_To_v1alpha1_AccountingConfig is an autogenerated conversion function.
func Convert_accounting_AccountingConfig_To_v1alpha1_AccountingConfig(in *accounting.AccountingConfig, out *AccountingConfig, s conversion.Scope) error {
	return autoConvert_accounting_AccountingConfig_To_v1alpha1_AccountingConfig(in, out, s)
}
