package machine

import (
	"fmt"
	"testing"

	gcpv1beta1 "github.com/openshift/cluster-api-provider-gcp/pkg/apis/gcpprovider/v1beta1"
	computeservice "github.com/openshift/cluster-api-provider-gcp/pkg/cloud/gcp/actuators/services/compute"
	machinev1beta1 "github.com/openshift/machine-api-operator/pkg/apis/machine/v1beta1"
	machinecontroller "github.com/openshift/machine-api-operator/pkg/controller/machine"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	controllerfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreate(t *testing.T) {
	_, mockComputeService := computeservice.NewComputeServiceMock()
	machineScope := machineScope{
		machine: &machinev1beta1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "",
				Namespace: "",
				Labels: map[string]string{
					machinev1beta1.MachineClusterIDLabel: "CLUSTERID",
				},
			},
		},
		coreClient:     controllerfake.NewFakeClient(),
		providerSpec:   &gcpv1beta1.GCPMachineProviderSpec{},
		providerStatus: &gcpv1beta1.GCPMachineProviderStatus{},
		computeService: mockComputeService,
	}
	reconciler := newReconciler(&machineScope)
	if err := reconciler.create(); err != nil {
		t.Errorf("reconciler was not expected to return error: %v", err)
	}
	if reconciler.providerStatus.Conditions[0].Type != gcpv1beta1.MachineCreated {
		t.Errorf("Expected: %s, got %s", gcpv1beta1.MachineCreated, reconciler.providerStatus.Conditions[0].Type)
	}
	if reconciler.providerStatus.Conditions[0].Status != corev1.ConditionTrue {
		t.Errorf("Expected: %s, got %s", corev1.ConditionTrue, reconciler.providerStatus.Conditions[0].Status)
	}
	if reconciler.providerStatus.Conditions[0].Reason != machineCreationSucceedReason {
		t.Errorf("Expected: %s, got %s", machineCreationSucceedReason, reconciler.providerStatus.Conditions[0].Reason)
	}
	if reconciler.providerStatus.Conditions[0].Message != machineCreationSucceedMessage {
		t.Errorf("Expected: %s, got %s", machineCreationSucceedMessage, reconciler.providerStatus.Conditions[0].Message)
	}
}

func TestReconcileMachineWithCloudState(t *testing.T) {
	_, mockComputeService := computeservice.NewComputeServiceMock()

	zone := "us-east1-b"
	projecID := "testProject"
	instanceName := "testInstance"
	machineScope := machineScope{
		machine: &machinev1beta1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instanceName,
				Namespace: "",
			},
		},
		coreClient: controllerfake.NewFakeClient(),
		providerSpec: &gcpv1beta1.GCPMachineProviderSpec{
			Zone: zone,
		},
		projectID:      projecID,
		providerID:     fmt.Sprintf("gce://%s/%s/%s", projecID, zone, instanceName),
		providerStatus: &gcpv1beta1.GCPMachineProviderStatus{},
		computeService: mockComputeService,
	}

	expectedNodeAddresses := []corev1.NodeAddress{
		{
			Type:    "InternalIP",
			Address: "10.0.0.15",
		},
		{
			Type:    "ExternalIP",
			Address: "35.243.147.143",
		},
	}

	r := newReconciler(&machineScope)
	if err := r.reconcileMachineWithCloudState(nil); err != nil {
		t.Errorf("reconciler was not expected to return error: %v", err)
	}
	if r.machine.Status.Addresses[0] != expectedNodeAddresses[0] {
		t.Errorf("Expected: %s, got: %s", expectedNodeAddresses[0], r.machine.Status.Addresses[0])
	}
	if r.machine.Status.Addresses[1] != expectedNodeAddresses[1] {
		t.Errorf("Expected: %s, got: %s", expectedNodeAddresses[1], r.machine.Status.Addresses[1])
	}

	if r.providerID != *r.machine.Spec.ProviderID {
		t.Errorf("Expected: %s, got: %s", r.providerID, *r.machine.Spec.ProviderID)
	}
	if *r.providerStatus.InstanceState != "RUNNING" {
		t.Errorf("Expected: %s, got: %s", "RUNNING", *r.providerStatus.InstanceState)
	}
	if *r.providerStatus.InstanceID != instanceName {
		t.Errorf("Expected: %s, got: %s", instanceName, *r.providerStatus.InstanceID)
	}
}

func TestExists(t *testing.T) {
	_, mockComputeService := computeservice.NewComputeServiceMock()
	machineScope := machineScope{
		machine: &machinev1beta1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "",
				Namespace: "",
				Labels: map[string]string{
					machinev1beta1.MachineClusterIDLabel: "CLUSTERID",
				},
			},
		},
		coreClient:     controllerfake.NewFakeClient(),
		providerSpec:   &gcpv1beta1.GCPMachineProviderSpec{},
		providerStatus: &gcpv1beta1.GCPMachineProviderStatus{},
		computeService: mockComputeService,
	}
	reconciler := newReconciler(&machineScope)
	exists, err := reconciler.exists()
	if err != nil || exists != true {
		t.Errorf("reconciler was not expected to return error: %v", err)
	}
}

func TestDelete(t *testing.T) {
	_, mockComputeService := computeservice.NewComputeServiceMock()
	machineScope := machineScope{
		machine: &machinev1beta1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "",
				Namespace: "",
				Labels: map[string]string{
					machinev1beta1.MachineClusterIDLabel: "CLUSTERID",
				},
			},
		},
		coreClient:     controllerfake.NewFakeClient(),
		providerSpec:   &gcpv1beta1.GCPMachineProviderSpec{},
		providerStatus: &gcpv1beta1.GCPMachineProviderStatus{},
		computeService: mockComputeService,
	}
	reconciler := newReconciler(&machineScope)
	if err := reconciler.delete(); err != nil {
		if _, ok := err.(*machinecontroller.RequeueAfterError); !ok {
			t.Errorf("reconciler was not expected to return error: %v", err)
		}
	}
}

func TestFmtInstanceSelfLink(t *testing.T) {
	expected := "https://www.googleapis.com/compute/v1/projects/a/zones/b/instances/c"
	res := fmtInstanceSelfLink("a", "b", "c")
	if res != expected {
		t.Errorf("Unexpected result from fmtInstanceSelfLink")
	}
}

type poolFuncTracker struct {
	called bool
}

func (p *poolFuncTracker) track(_, _ string) error {
	p.called = true
	return nil
}

func newPoolTracker() *poolFuncTracker {
	return &poolFuncTracker{
		called: false,
	}
}

func TestProcessTargetPools(t *testing.T) {
	_, mockComputeService := computeservice.NewComputeServiceMock()
	projecID := "testProject"
	instanceName := "testInstance"
	tpPresent := []string{
		"pool1",
	}
	tpEmpty := []string{}
	machineScope := machineScope{
		machine: &machinev1beta1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instanceName,
				Namespace: "",
			},
		},
		coreClient: controllerfake.NewFakeClient(),
		providerSpec: &gcpv1beta1.GCPMachineProviderSpec{
			Zone: "zone1",
		},
		projectID:      projecID,
		providerStatus: &gcpv1beta1.GCPMachineProviderStatus{},
		computeService: mockComputeService,
	}
	tCases := []struct {
		expectedCall bool
		desired      bool
		region       string
		targetPools  []string
	}{
		{
			// Delete when present
			expectedCall: true,
			desired:      false,
			region:       computeservice.WithMachineInPool,
			targetPools:  tpPresent,
		},
		{
			// Create when absent
			expectedCall: true,
			desired:      true,
			region:       computeservice.NoMachinesInPool,
			targetPools:  tpPresent,
		},
		{
			// Delete when absent
			expectedCall: false,
			desired:      false,
			region:       computeservice.NoMachinesInPool,
			targetPools:  tpPresent,
		},
		{
			// Create when present
			expectedCall: false,
			desired:      true,
			region:       computeservice.WithMachineInPool,
			targetPools:  tpPresent,
		},
		{
			// Return early when TP is empty list
			expectedCall: false,
			desired:      true,
			region:       computeservice.WithMachineInPool,
			targetPools:  tpEmpty,
		},
		{
			// Return early when TP is nil
			expectedCall: false,
			desired:      true,
			region:       computeservice.WithMachineInPool,
			targetPools:  nil,
		},
	}
	for i, tc := range tCases {
		pt := newPoolTracker()
		machineScope.providerSpec.Region = tc.region
		machineScope.providerSpec.TargetPools = tc.targetPools
		rec := newReconciler(&machineScope)
		err := rec.processTargetPools(tc.desired, pt.track)
		if err != nil {
			t.Errorf("unexpected error from ptp")
		}
		if pt.called != tc.expectedCall {
			t.Errorf("tc %v: expected didn't match observed: %v, %v", i, tc.expectedCall, pt.called)
		}
	}
}

func TestGetUserData(t *testing.T) {
	userDataSecretName := "test"
	defaultNamespace := "test"
	userDataBlob := "test"
	machineScope := machineScope{
		machine: &machinev1beta1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "",
				Namespace: defaultNamespace,
			},
		},
		providerSpec: &gcpv1beta1.GCPMachineProviderSpec{
			UserDataSecret: &corev1.LocalObjectReference{
				Name: userDataSecretName,
			},
		},
		providerStatus: &gcpv1beta1.GCPMachineProviderStatus{},
	}
	reconciler := newReconciler(&machineScope)

	testCases := []struct {
		secret *apiv1.Secret
		error  error
	}{
		{
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      userDataSecretName,
					Namespace: defaultNamespace,
				},
				Data: map[string][]byte{
					userDataSecretKey: []byte(userDataBlob),
				},
			},
			error: nil,
		},
		{
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "notFound",
					Namespace: defaultNamespace,
				},
				Data: map[string][]byte{
					userDataSecretKey: []byte(userDataBlob),
				},
			},
			error: &machinecontroller.MachineError{},
		},
		{
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      userDataSecretName,
					Namespace: defaultNamespace,
				},
				Data: map[string][]byte{
					"badKey": []byte(userDataBlob),
				},
			},
			error: &machinecontroller.MachineError{},
		},
	}

	for _, tc := range testCases {
		reconciler.coreClient = controllerfake.NewFakeClientWithScheme(scheme.Scheme, tc.secret)
		userData, err := reconciler.getCustomUserData()
		if tc.error != nil {
			if err == nil {
				t.Fatal("Expected error")
			}
			_, expectMachineError := tc.error.(*machinecontroller.MachineError)
			_, gotMachineError := err.(*machinecontroller.MachineError)
			if expectMachineError && !gotMachineError || !expectMachineError && gotMachineError {
				t.Errorf("Expected %T, got: %T", tc.error, err)
			}
		} else {
			if userData != userDataBlob {
				t.Errorf("Expected: %v, got: %v", userDataBlob, userData)
			}
		}
	}
}
