// Package v1beta1 is the v1beta1 version of the API.
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CronJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CronJobSpec `json:"spec"`
}

type CronJobSpec struct {
	Foo     string `json:"foo"`
	JobName string `json:"jobName"`
}
