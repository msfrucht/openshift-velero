/*
Copyright 2019 the Velero contributors.

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

package kube

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestParseResourceRequirements(t *testing.T) {
	type args struct {
		cpuRequest              string
		memRequest              string
		ephemeralStorageRequest string
		cpuLimit                string
		memLimit                string
		ephemeralStorageLimit   string
	}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		expectedError string
		expected      *corev1api.ResourceRequirements
	}{
		{
			name:          "unbounded quantities",
			args:          args{"0", "0", "0", "0", "0", "0"},
			wantErr:       false,
			expectedError: "",
			expected: &corev1api.ResourceRequirements{
				Requests: corev1api.ResourceList{},
				Limits:   corev1api.ResourceList{},
			},
		},
		{
			name:          "valid quantities",
			args:          args{"100m", "128Mi", "2Gi", "200m", "256Mi", "4Gi"},
			wantErr:       false,
			expectedError: "",
			expected: &corev1api.ResourceRequirements{
				Requests: corev1api.ResourceList{
					corev1api.ResourceCPU:              resource.MustParse("100m"),
					corev1api.ResourceMemory:           resource.MustParse("128Mi"),
					corev1api.ResourceEphemeralStorage: resource.MustParse("2Gi"),
				},
				Limits: corev1api.ResourceList{
					corev1api.ResourceCPU:              resource.MustParse("200m"),
					corev1api.ResourceMemory:           resource.MustParse("256Mi"),
					corev1api.ResourceEphemeralStorage: resource.MustParse("4Gi"),
				},
			},
		},
		{
			name:          "CPU request with unbounded limit",
			args:          args{"100m", "128Mi", "2Gi", "0", "256Mi", "2Gi"},
			wantErr:       false,
			expectedError: "",
			expected: &corev1api.ResourceRequirements{
				Requests: corev1api.ResourceList{
					corev1api.ResourceCPU:              resource.MustParse("100m"),
					corev1api.ResourceMemory:           resource.MustParse("128Mi"),
					corev1api.ResourceEphemeralStorage: resource.MustParse("2Gi"),
				},
				Limits: corev1api.ResourceList{
					corev1api.ResourceMemory:           resource.MustParse("256Mi"),
					corev1api.ResourceEphemeralStorage: resource.MustParse("2Gi"),
				},
			},
		},
		{
			name:          "Mem request with unbounded limit",
			args:          args{"100m", "128Mi", "2Gi", "200m", "0", "2Gi"},
			wantErr:       false,
			expectedError: "",
			expected: &corev1api.ResourceRequirements{
				Requests: corev1api.ResourceList{
					corev1api.ResourceCPU:              resource.MustParse("100m"),
					corev1api.ResourceMemory:           resource.MustParse("128Mi"),
					corev1api.ResourceEphemeralStorage: resource.MustParse("2Gi"),
				},
				Limits: corev1api.ResourceList{
					corev1api.ResourceCPU:              resource.MustParse("200m"),
					corev1api.ResourceEphemeralStorage: resource.MustParse("2Gi"),
				},
			},
		},
		{
			name:          "Ephemeral storage request with unbounded limit",
			args:          args{"100m", "128Mi", "2Gi", "200m", "128Mi", "0"},
			wantErr:       false,
			expectedError: "",
			expected: &corev1api.ResourceRequirements{
				Requests: corev1api.ResourceList{
					corev1api.ResourceCPU:              resource.MustParse("100m"),
					corev1api.ResourceMemory:           resource.MustParse("128Mi"),
					corev1api.ResourceEphemeralStorage: resource.MustParse("2Gi"),
				},
				Limits: corev1api.ResourceList{
					corev1api.ResourceCPU:    resource.MustParse("200m"),
					corev1api.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		},
		{
			name:          "CPU/Mem requests with unbounded limits",
			args:          args{"100m", "128Mi", "2Gi", "0", "0", "2Gi"},
			wantErr:       false,
			expectedError: "",
			expected: &corev1api.ResourceRequirements{
				Requests: corev1api.ResourceList{
					corev1api.ResourceCPU:              resource.MustParse("100m"),
					corev1api.ResourceMemory:           resource.MustParse("128Mi"),
					corev1api.ResourceEphemeralStorage: resource.MustParse("2Gi"),
				},
				Limits: corev1api.ResourceList{
					corev1api.ResourceEphemeralStorage: resource.MustParse("2Gi"),
				},
			},
		},
		{
			name:          "CPU/Mem/ephemeral-storage requests with unbounded limits",
			args:          args{"100m", "128Mi", "2Gi", "0", "0", "0"},
			wantErr:       false,
			expectedError: "",
			expected: &corev1api.ResourceRequirements{
				Requests: corev1api.ResourceList{
					corev1api.ResourceCPU:              resource.MustParse("100m"),
					corev1api.ResourceMemory:           resource.MustParse("128Mi"),
					corev1api.ResourceEphemeralStorage: resource.MustParse("2Gi"),
				},
				Limits: corev1api.ResourceList{},
			},
		},
		{
			name:          "invalid quantity",
			args:          args{"100m", "invalid", "2Gi", "200m", "256Mi", "2Gi"},
			wantErr:       true,
			expectedError: `couldn't parse memory request "invalid": quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'`,
			expected:      nil,
		},
		{
			name:          "CPU request greater than limit",
			args:          args{"300m", "128Mi", "2Gi", "200m", "256Mi", "2Gi"},
			wantErr:       true,
			expectedError: `CPU request "300m" must be less than or equal to CPU limit "200m"`,
			expected:      nil,
		},
		{
			name:          "memory request greater than limit",
			args:          args{"100m", "512Mi", "2Gi", "200m", "256Mi", "2Gi"},
			wantErr:       true,
			expectedError: `Memory request "512Mi" must be less than or equal to Memory limit "256Mi"`,
			expected:      nil,
		},
		{
			name:          "Ephemeral storage request greater than limit",
			args:          args{"100m", "128Mi", "4Gi", "200m", "256Mi", "2Gi"},
			wantErr:       true,
			expectedError: `Ephemeral storage request "4Gi" must be less than or equal to ephemeral storage limit "2Gi"`,
			expected:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseResourceRequirements(tt.args.cpuRequest, tt.args.memRequest, tt.args.ephemeralStorageRequest, tt.args.cpuLimit, tt.args.memLimit, tt.args.ephemeralStorageLimit)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}
			require.NoError(t, err)

			var expected corev1api.ResourceRequirements
			if tt.expected == nil {
				expected = corev1api.ResourceRequirements{
					Requests: corev1api.ResourceList{
						corev1api.ResourceCPU:              resource.MustParse(tt.args.cpuRequest),
						corev1api.ResourceMemory:           resource.MustParse(tt.args.memRequest),
						corev1api.ResourceEphemeralStorage: resource.MustParse(tt.args.ephemeralStorageRequest),
					},
					Limits: corev1api.ResourceList{
						corev1api.ResourceCPU:              resource.MustParse(tt.args.cpuLimit),
						corev1api.ResourceMemory:           resource.MustParse(tt.args.memLimit),
						corev1api.ResourceEphemeralStorage: resource.MustParse(tt.args.ephemeralStorageLimit),
					},
				}
			} else {
				expected = *tt.expected
			}

			assert.Equal(t, expected, got)
		})
	}
}
