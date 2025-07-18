/*
Copyright The Kubernetes Authors.

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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1beta1

import (
	appsv1beta1 "k8s.io/api/apps/v1beta1"
)

// DeploymentStrategyApplyConfiguration represents a declarative configuration of the DeploymentStrategy type for use
// with apply.
type DeploymentStrategyApplyConfiguration struct {
	Type          *appsv1beta1.DeploymentStrategyType        `json:"type,omitempty"`
	RollingUpdate *RollingUpdateDeploymentApplyConfiguration `json:"rollingUpdate,omitempty"`
}

// DeploymentStrategyApplyConfiguration constructs a declarative configuration of the DeploymentStrategy type for use with
// apply.
func DeploymentStrategy() *DeploymentStrategyApplyConfiguration {
	return &DeploymentStrategyApplyConfiguration{}
}
func (b DeploymentStrategyApplyConfiguration) IsApplyConfiguration() {}

// WithType sets the Type field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Type field is set to the value of the last call.
func (b *DeploymentStrategyApplyConfiguration) WithType(value appsv1beta1.DeploymentStrategyType) *DeploymentStrategyApplyConfiguration {
	b.Type = &value
	return b
}

// WithRollingUpdate sets the RollingUpdate field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RollingUpdate field is set to the value of the last call.
func (b *DeploymentStrategyApplyConfiguration) WithRollingUpdate(value *RollingUpdateDeploymentApplyConfiguration) *DeploymentStrategyApplyConfiguration {
	b.RollingUpdate = value
	return b
}
