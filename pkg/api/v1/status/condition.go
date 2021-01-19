package status

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*

conditions:
  - lastTransitionTime: "2020-12-15T20:46:55Z"
    status: "True"
    message: the following fields are missing in the Secret secret: %v
    reason: AtlasCredentialsNotProvided
    type: ProjectReady
  - lastTransitionTime: "2020-12-15T20:46:55Z"
    message: NOT_ALLOWED. You don't have enough permissions to perform the operation
    reason: AtlasApiError
    status: "False"
    type: IPAccessListReady
  - privateLink
  - lastTransitionTime: "2020-12-15T20:46:55Z"
    status: "True"
    type: Ready

*/

type ConditionType string

const (
	ReadyType ConditionType = "Ready"
)

// AtlasProject condition types
const (
	ProjectReadyType      ConditionType = "ProjectReady"
	IPAccessListReadyType ConditionType = "IPAccessListReady"
)

// AtlasCluster condition types
const (
	ClusterReadyType ConditionType = "ClusterReady"
)

// Condition describes the state of an Atlas Custom Resource at a certain point.
type Condition struct {
	// Type of Atlas Custom Resource condition.
	Type ConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// TrueCondition returns the Condition that has the 'Status' set to 'true' and 'Type' to 'conditionType'.
// It explicitly omits the 'Reason' and 'Message' fields.
func TrueCondition(conditionType ConditionType) Condition {
	return Condition{
		Type:               conditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
}

// FalseCondition returns the COndition that has the 'Status' set to 'false' and 'Type' to 'conditionType'.
// The reason and message can be provided optionally
func FalseCondition(conditionType ConditionType, reasonAndMessage ...string) Condition {
	condition := Condition{
		Type:               conditionType,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
	}
	if len(reasonAndMessage) >= 1 {
		condition.Reason = reasonAndMessage[0]
	}
	if len(reasonAndMessage) == 2 {
		condition.Message = reasonAndMessage[1]
	}
	return condition
}