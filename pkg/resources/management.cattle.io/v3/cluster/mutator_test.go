package cluster

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/webhook/pkg/admission"
	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	rketypes "github.com/rancher/rke/types"
)

func TestManagementClusterMutator_Admit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		oldCluster    unstructured.Unstructured
		newCluster    *v3.Cluster
		operation     admissionv1.Operation
		expectAllowed bool
	}{
		{
			name:          "Create",
			operation:     admissionv1.Create,
			expectAllowed: true,
		},
		{
			name:      "Create an RKE1",
			operation: admissionv1.Create,
			newCluster: &v3.Cluster{
				Spec: v3.ClusterSpec{
					ClusterSpecBase: v3.ClusterSpecBase{
						RancherKubernetesEngineConfig: &rketypes.RancherKubernetesEngineConfig{
							Network: rketypes.NetworkConfig{
								Options: map[string]string{"key": "value"},
							},
						},
					},
				},
			},
			oldCluster: unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": "management.cattle.io/v3",
				"kind":       "Cluster",
				"metadata": map[string]interface{}{
					"name": "c-4r26s",
				},
				"spec": map[string]interface{}{
					"rancherKubernetesEngineConfig": map[string]interface{}{
						"network": map[string]interface{}{
							"options": map[string]interface{}{
								"flannel_backend_type": "vxlan",
							},
							"plugin":       "canal",
							"nosuchoption": "somevalue",
						},
					},
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ManagementClusterMutator{}

			oldClusterBytes, err := json.Marshal(tt.oldCluster)
			assert.NoError(t, err)
			newClusterBytes, err := json.Marshal(tt.newCluster)
			assert.NoError(t, err)

			resp, err := m.Admit(&admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: newClusterBytes,
					},
					OldObject: runtime.RawExtension{
						Raw: oldClusterBytes,
					},
					Operation: tt.operation,
				},
			})
			assert.NoError(t, err)
			assert.Equal(t, tt.expectAllowed, resp.Allowed)
			fmt.Print(resp.Allowed)
		})
	}
}
