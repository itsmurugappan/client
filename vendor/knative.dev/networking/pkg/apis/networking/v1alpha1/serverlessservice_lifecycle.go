/*
Copyright 2019 The Knative Authors.

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

package v1alpha1

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/pkg/apis"
)

var serverlessServiceCondSet = apis.NewLivingConditionSet(
	ServerlessServiceConditionEndspointsPopulated,
)

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*ServerlessService) GetConditionSet() apis.ConditionSet {
	return serverlessServiceCondSet
}

// GetGroupVersionKind returns the GVK for the ServerlessService.
func (ss *ServerlessService) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ServerlessService")
}

// GetCondition returns the value of the condition `t`.
func (sss *ServerlessServiceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return serverlessServiceCondSet.Manage(sss).GetCondition(t)
}

// InitializeConditions initializes the conditions.
func (sss *ServerlessServiceStatus) InitializeConditions() {
	serverlessServiceCondSet.Manage(sss).InitializeConditions()
}

// MarkEndpointsReady marks the ServerlessServiceStatus endpoints populated condition to true.
func (sss *ServerlessServiceStatus) MarkEndpointsReady() {
	serverlessServiceCondSet.Manage(sss).MarkTrue(ServerlessServiceConditionEndspointsPopulated)
}

// MarkEndpointsNotOwned marks that we don't own K8s service.
func (sss *ServerlessServiceStatus) MarkEndpointsNotOwned(kind, name string) {
	serverlessServiceCondSet.Manage(sss).MarkFalse(
		ServerlessServiceConditionEndspointsPopulated, "NotOwned",
		"Resource %s of type %s is not owned by SKS", name, kind)
}

// MarkActivatorEndpointsPopulated is setting the ActivatorEndpointsPopulated to True.
func (sss *ServerlessServiceStatus) MarkActivatorEndpointsPopulated() {
	serverlessServiceCondSet.Manage(sss).SetCondition(apis.Condition{
		Type:     ActivatorEndpointsPopulated,
		Status:   corev1.ConditionTrue,
		Severity: apis.ConditionSeverityInfo,
		Reason:   "ActivatorEndpointsPopulated",
		Message:  "Revision is backed by Activator",
	})
}

// MarkActivatorEndpointsRemoved is setting the ActivatorEndpointsPopulated to False.
func (sss *ServerlessServiceStatus) MarkActivatorEndpointsRemoved() {
	serverlessServiceCondSet.Manage(sss).SetCondition(apis.Condition{
		Type:     ActivatorEndpointsPopulated,
		Status:   corev1.ConditionFalse,
		Severity: apis.ConditionSeverityInfo,
		Reason:   "ActivatorEndpointsPopulated",
		Message:  "Revision is backed by Activator",
	})
}

// MarkEndpointsNotReady marks the ServerlessServiceStatus endpoints populated condition to unknown.
func (sss *ServerlessServiceStatus) MarkEndpointsNotReady(reason string) {
	serverlessServiceCondSet.Manage(sss).MarkUnknown(
		ServerlessServiceConditionEndspointsPopulated, reason,
		"K8s Service is not ready")
}

// IsReady returns true if the Status condition Ready
// is true and the latest spec has been observed.
func (ss *ServerlessService) IsReady() bool {
	sss := ss.Status
	return sss.ObservedGeneration == ss.Generation &&
		sss.GetCondition(ServerlessServiceConditionReady).IsTrue()
}

// ProxyFor returns how long it has been since Activator was moved
// to the request path.
func (sss *ServerlessServiceStatus) ProxyFor() time.Duration {
	cond := sss.GetCondition(ActivatorEndpointsPopulated)
	if cond == nil || cond.Status != corev1.ConditionTrue {
		return 0
	}
	return time.Since(cond.LastTransitionTime.Inner.Time)
}
