/*
Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.

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

package clientwrapper

import (
	"context"
	"strings"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/pkg/errors"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/applicationautoscalingiface"
)

// Provides the prefixes and error codes relating to each endpoint
// TODO mbaijal: Check the messages
const (
	DescribeHAP404MessagePrefix = "Could not find HAP"
	DescribeHAP404Code          = "ValidationException"
	DeleteHAP404MessagePrefix   = "Could not find HAP"
	DeleteHAP404Code            = "ObjectNotFoundException"
	HAP500MessagePrefix   		= "Server Error"
	HAP500Code            		= "InternalServiceException"
)

// ApplicationAutoscalingClientWrapper interface for ApplicationAutoscalingClient wrapper
type ApplicationAutoscalingClientWrapper interface {
	RegisterScalableTarget(ctx context.Context, autoscalingTarget *applicationautoscaling.RegisterScalableTargetInput) (*applicationautoscaling.RegisterScalableTargetOutput, error)
	PutScalingPolicy(ctx context.Context, autoscalingJob *applicationautoscaling.PutScalingPolicyInput) (*applicationautoscaling.PutScalingPolicyOutput, error)
	DeleteScalingPolicy(ctx context.Context, autoscalingJob *applicationautoscaling.DeleteScalingPolicyInput) (*applicationautoscaling.DeleteScalingPolicyOutput, error)
	DeregisterScalableTarget(ctx context.Context, autoscalingJob *applicationautoscaling.DeregisterScalableTargetInput) (*applicationautoscaling.DeregisterScalableTargetOutput, error)
	DescribeScalableTargets(ctx context.Context, resourceID string) (*applicationautoscaling.DescribeScalableTargetsOutput, error)
	DescribeScalingPolicies(ctx context.Context, policyName string, resourceID string) (*applicationautoscaling.ScalingPolicy, error)
}

// NewApplicationAutoscalingClientWrapper creates a ApplicationAutoscaling wrapper around an existing client.
func NewApplicationAutoscalingClientWrapper(innerClient applicationautoscalingiface.ClientAPI) ApplicationAutoscalingClientWrapper {
	return &applicationAutoscalingClientWrapper{
		innerClient: innerClient,
	}
}

// ApplicationAutoscalingClientWrapperProvider defines a function that returns a ApplicationAutoscaling client. Used for mocking.
type ApplicationAutoscalingClientWrapperProvider func(aws.Config) ApplicationAutoscalingClientWrapper

// Implementation of ApplicationAutoscaling client wrapper.
type applicationAutoscalingClientWrapper struct {
	ApplicationAutoscalingClientWrapper
	innerClient applicationautoscalingiface.ClientAPI
}

// RegisterScalableTarget registers a scalable target. Returns the response output or nil if error.
func (c *applicationAutoscalingClientWrapper) RegisterScalableTarget(ctx context.Context, autoscalingTarget *applicationautoscaling.RegisterScalableTargetInput) (*applicationautoscaling.RegisterScalableTargetOutput, error) {

	createRequest := c.innerClient.RegisterScalableTargetRequest(autoscalingTarget)
	response, err := createRequest.Send(ctx)

	if response != nil {
		return response.RegisterScalableTargetOutput, nil
	}
	return nil, err
}

// RegisterScalableTarget registers a scalable target. Returns the response output or nil if error.
func (c *applicationAutoscalingClientWrapper) PutScalingPolicy(ctx context.Context, autoscalingJob *applicationautoscaling.PutScalingPolicyInput) (*applicationautoscaling.PutScalingPolicyOutput, error) {

	createRequest := c.innerClient.PutScalingPolicyRequest(autoscalingJob)
	response, err := createRequest.Send(ctx)

	if response != nil {
		return response.PutScalingPolicyOutput, nil
	}

	return nil, err
}

// DeleteScalingPolicy Deletes the scaling policy
func (c *applicationAutoscalingClientWrapper) DeleteScalingPolicy(ctx context.Context, autoscalingJob *applicationautoscaling.DeleteScalingPolicyInput) (*applicationautoscaling.DeleteScalingPolicyOutput, error) {
	deleteRequest := c.innerClient.DeleteScalingPolicyRequest(autoscalingJob)
	deleteResponse, deleteError := deleteRequest.Send(ctx)

	if deleteError != nil {
		return nil, deleteError
	}

	return deleteResponse.DeleteScalingPolicyOutput, deleteError
}

// DeregisterScalableTarget deregisters a scalable target
func (c *applicationAutoscalingClientWrapper) DeregisterScalableTarget(ctx context.Context, autoscalingJob *applicationautoscaling.DeregisterScalableTargetInput) (*applicationautoscaling.DeregisterScalableTargetOutput, error) {
	deleteRequest := c.innerClient.DeregisterScalableTargetRequest(autoscalingJob)
	deleteResponse, deleteError := deleteRequest.Send(ctx)

	if deleteError != nil {
		return nil, deleteError
	}

	return deleteResponse.DeregisterScalableTargetOutput, deleteError
}

// DescribeScalableTargets returns the scalableTarget description filtered on PolicyName and a single ResourceID
// TODO: change this to return only the ScalableTargetObject for cleaner descriptions
func (c *applicationAutoscalingClientWrapper) DescribeScalableTargets(ctx context.Context, resourceID string) (*applicationautoscaling.DescribeScalableTargetsOutput, error) {

	var resourceIDList []string
	resourceIDList = append(resourceIDList, resourceID)
	// Review: This filtered response should be of size 1 by default
	var maxResults int64 = 1

	// TODO: Remove hardcoded values, might need to construct the input object
	describeRequest := c.innerClient.DescribeScalableTargetsRequest(&applicationautoscaling.DescribeScalableTargetsInput{
		ResourceIds:       resourceIDList,
		MaxResults:        &maxResults,
		ScalableDimension: "sagemaker:variant:DesiredInstanceCount",
		ServiceNamespace:  "sagemaker",
	})

	describeResponse, describeError := describeRequest.Send(ctx)

	if describeError != nil {
		return nil, describeError
	}

	return describeResponse.DescribeScalableTargetsOutput, describeError
}

// DescribeScalingPolicies returns the scaling policy description filtered on PolicyName and a single ResourceID
// returns only the scalingPolicy object else the actionDetermination gets messy
func (c *applicationAutoscalingClientWrapper) DescribeScalingPolicies(ctx context.Context, policyName string, resourceID string) (*applicationautoscaling.ScalingPolicy, error) {

	var policyNameList []string
	var scalingPolicyDescription *applicationautoscaling.ScalingPolicy
	policyNameList = append(policyNameList, policyName)
	// Review: This filtered response should be of size 1 by default
	var maxResults int64 = 1

	// TODO: Remove hardcoded values, might need to construct the inputs
	describeRequest := c.innerClient.DescribeScalingPoliciesRequest(&applicationautoscaling.DescribeScalingPoliciesInput{
		PolicyNames:       policyNameList,
		MaxResults:        &maxResults,
		ResourceId:        &resourceID,
		ScalableDimension: "sagemaker:variant:DesiredInstanceCount",
		ServiceNamespace:  "sagemaker",
	})

	describeResponse, describeError := describeRequest.Send(ctx)

	if describeError != nil {
		return scalingPolicyDescription, describeError
	}
	
	// Review: Slightly Hacky, but valid
	if len(describeResponse.DescribeScalingPoliciesOutput.ScalingPolicies) == 1 {
		scalingPolicyDescription = &(describeResponse.DescribeScalingPoliciesOutput.ScalingPolicies[0])
	} else {
		scalingPolicyDescription = nil
	}

	return scalingPolicyDescription, describeError
}

// IsDeleteHAP404Error determines whether the given error is equivalent to an HTTP 404 status code.
func IsDeleteHAP404Error(err error) bool {
	awserror := errors.Cause(err)
	if requestFailure, isRequestFailure := awserror.(awserr.RequestFailure); isRequestFailure {
		return requestFailure.Code() == DeleteModel404Code && strings.HasPrefix(requestFailure.Message(), DeleteHAP404MessagePrefix)
	}

	return false
}

// IsDescribeHAP404Error detects if the error is a 404
func IsDescribeHAP404Error(err error) bool {
	awserror := errors.Cause(err)
	if requestFailure, isRequestFailure := awserror.(awserr.RequestFailure); isRequestFailure {
		return requestFailure.Code() == DescribeHAP404Code && strings.HasPrefix(requestFailure.Message(), DescribeHAP404MessagePrefix)
	} 
	return false
}

// IsHAP500Error detects if the error is a 404
func IsHAP500Error(err error) bool {
	awserror := errors.Cause(err)
	if requestFailure, isRequestFailure := awserror.(awserr.RequestFailure); isRequestFailure {
		return requestFailure.Code() == HAP500Code && strings.HasPrefix(requestFailure.Message(), HAP500MessagePrefix)
	} 
	return false
}
