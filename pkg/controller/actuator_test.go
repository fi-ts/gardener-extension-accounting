package controller

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

func TestDecodeInfrastructureConfig(t *testing.T) {
	tests := []struct {
		name    string
		raw     *runtime.RawExtension
		wantErr bool
	}{
		{
			name: "unknown fields are ignored",
			raw: &runtime.RawExtension{Raw: []byte(`{
				"apiVersion": "metal.gardener.cloud/v1alpha1",
				"kind": "InfrastructureConfig",
				"partitionID": "partition-a",
				"projectID": "project-a",
				"firewall": {
					"size": "c1-small",
					"suricata": true
				}
			}`)},
		},
		{
			name:    "missing projectID",
			raw:     &runtime.RawExtension{Raw: []byte(`{"partitionID": "partition-a"}`)},
			wantErr: true,
		},
		{
			name:    "missing partitionID",
			raw:     &runtime.RawExtension{Raw: []byte(`{"projectID": "project-a"}`)},
			wantErr: true,
		},
		{
			name:    "missing provider config",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := decodeInfrastructureConfig(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if config.ProjectID != "project-a" {
				t.Errorf("expected projectID %q, got %q", "project-a", config.ProjectID)
			}
			if config.PartitionID != "partition-a" {
				t.Errorf("expected partitionID %q, got %q", "partition-a", config.PartitionID)
			}
		})
	}
}
