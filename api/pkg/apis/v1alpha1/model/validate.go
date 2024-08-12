/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type (
	ErrorField struct {
		FieldPath       string
		Value           interface{}
		DetailedMessage string
	}
	IValidation interface {
		ValidateCreateOrUpdate(ctx context.Context, old IValidation) []ErrorField
		ValidateDelete(ctx context.Context) []ErrorField
	}

	ResourceType string

	// Prototype for object lookup functions. Return value indicates if the object exists or not.
	ObjectLookupFunc func(ctx context.Context, objectName string, namespace string) (interface{}, error)
	// Prototype for linked objects lookup functions.
	LinkedObjectLookupFunc func(ctx context.Context, objectName string, namespace string) (bool, error)
)

const (
	Target            ResourceType = "target"
	Device            ResourceType = "device"
	Solution          ResourceType = "solution"
	Instance          ResourceType = "instance"
	Campaign          ResourceType = "campaign"
	Activation        ResourceType = "activation"
	Catalog           ResourceType = "catalog"
	SolutionContainer ResourceType = "solutioncontainer"
	CampaignContainer ResourceType = "campaigncontainer"
	CatalogContainer  ResourceType = "catalogcontainer"
)

func GetResourceMetadata(resourceType ResourceType) (string, string, string, string) {
	var group string
	var version string
	var resource string
	var kind string
	switch resourceType {
	case Solution:
		group = "solution.symphony"
		version = "v1"
		resource = "solutions"
		kind = "Solution"
	case SolutionContainer:
		group = "solution.symphony"
		version = "v1"
		resource = "solutioncontainers"
		kind = "SolutionContainer"
	case Instance:
		group = "solution.symphony"
		version = "v1"
		resource = "instances"
		kind = "Instance"
	case Target:
		group = "fabric.symphony"
		version = "v1"
		resource = "targets"
		kind = "Target"
	case Device:
		group = "fabric.symphony"
		version = "v1"
		resource = "devices"
		kind = "Device"
	case Campaign:
		group = "workflow.symphony"
		version = "v1"
		resource = "campaigns"
		kind = "Campaign"
	case CampaignContainer:
		group = "workflow.symphony"
		version = "v1"
		resource = "campaigncontainers"
		kind = "CampaignContainer"
	case Activation:
		group = "workflow.symphony"
		version = "v1"
		resource = "activations"
		kind = "Activation"
	case Catalog:
		group = "federation.symphony"
		version = "v1"
		resource = "catalogs"
		kind = "Catalog"
	case CatalogContainer:
		group = "federation.symphony"
		version = "v1"
		resource = "catalogcontainers"
		kind = "CatalogContainer"
	default:
		group = ""
		version = ""
		resource = ""
		kind = ""
	}
	return group, version, resource, kind
}

func ConvertReferenceToObjectName(name string) string {
	if strings.Contains(name, constants.ReferenceSeparator) {
		name = strings.ReplaceAll(name, constants.ReferenceSeparator, constants.ResourceSeperator)
	}
	return name
}

func ConvertObjectNameToReference(name string) string {
	index := strings.LastIndex(name, constants.ResourceSeperator)
	if index == -1 {
		return name
	}
	return name[:index] + constants.ReferenceSeparator + name[index+len(constants.ResourceSeperator):]
}

func ConvertErrorFieldsToK8sError(ErrorFields []ErrorField) field.ErrorList {
	var allErrs field.ErrorList
	for _, errorField := range ErrorFields {
		pathArray := strings.Split(errorField.FieldPath, ".")
		errorPath := field.NewPath(pathArray[0])
		for _, path := range pathArray[1:] {
			errorPath = errorPath.Child(path)
		}
		allErrs = append(allErrs, field.Invalid(errorPath, errorField.Value, errorField.DetailedMessage))
	}
	return allErrs
}

func ConvertErrorFieldsToString(ErrorFields []ErrorField) string {
	errorMessages := ""
	for _, errorField := range ErrorFields {
		errorMessage := errorField.FieldPath + ": " + "Invalid value: " + utils.FormatAsString(errorField.Value) + ": " + errorField.DetailedMessage
		errorMessages = errorMessages + errorMessage + "\n"
	}

	return errorMessages
}

func ValidateCreateOrUpdate(ctx context.Context, newObj IValidation, oldObj IValidation, errorWhenGetOldObj error) error {
	var errorFields []ErrorField
	if errorWhenGetOldObj != nil {
		if v1alpha2.IsNotFound(errorWhenGetOldObj) {
			errorFields = newObj.ValidateCreateOrUpdate(ctx, nil)
		} else {
			return v1alpha2.NewCOAError(errorWhenGetOldObj, "Unable to get previous state from state store when validating the create or update request", v1alpha2.InternalError)
		}
	} else {
		errorFields = newObj.ValidateCreateOrUpdate(ctx, oldObj)
	}
	if len(errorFields) > 0 {
		errorMessage := ConvertErrorFieldsToString(errorFields)
		return v1alpha2.NewCOAError(nil, "Failed to create or update object: "+errorMessage, v1alpha2.BadRequest)
	} else {
		return nil
	}
}

func ValidateDelete(ctx context.Context, obj IValidation, errorWhenGetObj error) error {
	if errorWhenGetObj != nil {
		if v1alpha2.IsNotFound(errorWhenGetObj) {
			return nil
		} else {
			return v1alpha2.NewCOAError(errorWhenGetObj, "Unable to get current state from state store when validating the delete request", v1alpha2.InternalError)
		}
	} else {
		errorFields := obj.ValidateDelete(ctx)
		if len(errorFields) > 0 {
			errorMessage := ConvertErrorFieldsToString(errorFields)
			return v1alpha2.NewCOAError(nil, "Failed to delete instance: "+errorMessage, v1alpha2.BadRequest)
		} else {
			return nil
		}
	}
}

// rootResource is not in metadata now, pass in as a parameter
func (o *ObjectMeta) ValidateRootResource(ctx context.Context, rootResource string, lookupFunc ObjectLookupFunc) *ErrorField {
	if lookupFunc == nil {
		return nil
	}
	if _, err := lookupFunc(ctx, rootResource, o.Namespace); err != nil {
		return &ErrorField{
			FieldPath:       "spec.rootResource",
			Value:           rootResource,
			DetailedMessage: "rootResource must be a valid container",
		}
	}
	// ownerreferences check is only appliable to k8s
	return nil
}

func ValidateObjectName(name string, rootResource string) *ErrorField {
	if rootResource == "" {
		return &ErrorField{
			FieldPath:       "spec.rootResource",
			Value:           rootResource,
			DetailedMessage: "rootResource must be a non-empty string",
		}
	}

	if !strings.HasPrefix(name, rootResource) {
		return &ErrorField{
			FieldPath:       "metadata.name",
			Value:           name,
			DetailedMessage: "name must start with spec.rootResource",
		}
	}

	prefix := rootResource + constants.ResourceSeperator
	remaining := strings.TrimPrefix(name, prefix)

	if remaining == name {
		return &ErrorField{
			FieldPath:       "metadata.name",
			Value:           name,
			DetailedMessage: fmt.Sprintf("name should be in the format '<rootResource>%s<version>'", constants.ResourceSeperator),
		}

	}

	if strings.Contains(remaining, constants.ResourceSeperator) || strings.HasPrefix(remaining, "v-") {
		return &ErrorField{
			FieldPath:       "metadata.name",
			Value:           name,
			DetailedMessage: "name should be in the format <rootResource>-v-<version> where <version> does not contain '-v-' or starts with 'v-'",
		}
	}

	return nil
}
