/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package validation

import (
	"context"
	"encoding/json"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
)

var (
	catalogversionMaxNameLength = 61
	catalogversionMinNameLength = 1
)

type CatalogVersionValidator struct {
	// Check CatalogVersion existence
	CatalogVersionLookupFunc          ObjectLookupFunc
	CatalogLookupFunc ObjectLookupFunc
	ChildCatalogVersionLookupFunc     LinkedObjectLookupFunc
}

func NewCatalogVersionValidator(catalogversionLookupFunc ObjectLookupFunc, catalogLookupFunc ObjectLookupFunc, childCatalogVersionLookupFunc LinkedObjectLookupFunc) CatalogVersionValidator {
	return CatalogVersionValidator{
		CatalogVersionLookupFunc:          catalogversionLookupFunc,
		CatalogLookupFunc: catalogLookupFunc,
		ChildCatalogVersionLookupFunc:     childCatalogVersionLookupFunc,
	}
}

// Validate CatalogVersion creation or update
// 1. Schema is valid
// 2. Parent catalogversion exists
// 3. CatalogVersion name and rootResource is valid. And rootResource is immutable
// TODO: 4. Update won't form a cycle in the parent-child relationship
func (c *CatalogVersionValidator) ValidateCreateOrUpdate(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := c.ConvertInterfaceToCatalogVersion(newRef)
	old := c.ConvertInterfaceToCatalogVersion(oldRef)

	errorFields := []ErrorField{}
	if err := c.ValidateSchema(ctx, new); err != nil {
		errorFields = append(errorFields, *err)
	}
	if new.Spec.ParentName != "" && (oldRef == nil || new.Spec.ParentName != old.Spec.ParentName) {
		if err := c.ValidateParentCatalogVersion(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}

	if oldRef == nil {
		// validate create specific fields
		if err := ValidateObjectName(new.ObjectMeta.Name, new.Spec.RootResource, catalogversionMinNameLength, catalogversionMaxNameLength); err != nil {
			errorFields = append(errorFields, *err)
		}
		// validate rootResource
		if err := ValidateRootResource(ctx, new.ObjectMeta, new.Spec.RootResource, c.CatalogLookupFunc); err != nil {
			errorFields = append(errorFields, *err)
		}
	} else {
		// validate update specific fields
		if new.Spec.RootResource != old.Spec.RootResource {
			errorFields = append(errorFields, ErrorField{
				FieldPath:       "spec.rootResource",
				Value:           new.Spec.RootResource,
				DetailedMessage: "rootResource is immutable",
			})
		}
	}

	return errorFields
}

// Validate CatalogVersion deletion
// 1. CatalogVersion has no child catalogversions
func (c *CatalogVersionValidator) ValidateDelete(ctx context.Context, catalogversion interface{}) []ErrorField {
	new := c.ConvertInterfaceToCatalogVersion(catalogversion)
	// validate child catalogversions
	errorFields := []ErrorField{}
	if err := c.ValidateChildCatalogVersion(ctx, new); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

// Validate Schema is valid
func (c *CatalogVersionValidator) ValidateSchema(ctx context.Context, new model.CatalogVersionState) *ErrorField {
	if c.CatalogVersionLookupFunc == nil {
		return nil
	}
	if schemaName, ok := new.Spec.Metadata["schema"]; ok {
		// 1). Lookup catalogversion object with schema name
		schemaName = ConvertReferenceToObjectName(schemaName)
		lookupRes, err := c.CatalogVersionLookupFunc(ctx, schemaName, new.ObjectMeta.Namespace)

		if err != nil {
			return &ErrorField{
				FieldPath:       "spec.metadata.schema",
				Value:           schemaName,
				DetailedMessage: "could not find the required schema",
			}
		}
		marshalResult, _ := json.Marshal(lookupRes)
		var catalogversion model.CatalogVersionState
		err = json.Unmarshal(marshalResult, &catalogversion)
		if err != nil {
			return &ErrorField{
				FieldPath:       "spec.metadata.schema",
				Value:           schemaName,
				DetailedMessage: "schema is not a valid catalogversion object",
			}
		}
		if spec, ok := catalogversion.Spec.Properties["spec"]; ok {
			// 2). Extract Schema object from the catalogversion object
			var schemaObj utils.Schema
			jData, _ := json.Marshal(spec)
			err := json.Unmarshal(jData, &schemaObj)
			if err != nil {
				return &ErrorField{
					FieldPath:       "spec.metadata.schema",
					Value:           schemaName,
					DetailedMessage: "invalid schema",
				}
			}

			// 3). Validate the schema on the catalogversion which is being created/updated
			result, err := schemaObj.CheckProperties(ctx, new.Spec.Properties, nil)
			if err != nil {
				return &ErrorField{
					FieldPath:       "spec.metadata.schema",
					Value:           schemaName,
					DetailedMessage: "unable to determine the validity of the schema",
				}
			}
			if !result.Valid {
				return &ErrorField{
					FieldPath:       "spec.Properties",
					Value:           "(hidden)",
					DetailedMessage: "invalid schema result: " + result.ToErrorMessages(),
				}
			}
		}
	}
	return nil
}

// Validate Parent CatalogVersion exists if provided
// Use CatalogVersionLookupFunc to lookup the catalogversion with parentName
func (c *CatalogVersionValidator) ValidateParentCatalogVersion(ctx context.Context, catalogversion model.CatalogVersionState) *ErrorField {
	if c.CatalogVersionLookupFunc == nil {
		return nil
	}
	parentCatalogVersionName := ConvertReferenceToObjectName(catalogversion.Spec.ParentName)
	parentCatalogVersion, err := c.CatalogVersionLookupFunc(ctx, parentCatalogVersionName, catalogversion.ObjectMeta.Namespace)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.ParentName",
			Value:           parentCatalogVersionName,
			DetailedMessage: "parent catalogversion not found",
		}
	}

	// if parent exists, need to check if upsert this catalogversion will introduce circular parent issue.
	catalogversionName := catalogversion.ObjectMeta.Name
	if c.hasParentCircularDependency(ctx, parentCatalogVersion, catalogversionName) {
		return &ErrorField{
			FieldPath:       "spec.ParentName",
			Value:           parentCatalogVersionName,
			DetailedMessage: "parent catalogversion has circular dependency",
		}
	}
	return nil
}

func (c *CatalogVersionValidator) hasParentCircularDependency(ctx context.Context, parentRaw interface{}, catalogversionName string) bool {
	jsonData, err := json.Marshal(parentRaw)
	if err != nil {
		return false
	}

	var parentCatalogVersion model.CatalogVersionState
	err = json.Unmarshal(jsonData, &parentCatalogVersion)
	if err != nil {
		return false
	}

	if parentCatalogVersion.Spec.ParentName == "" {
		return false
	} else {
		parentName := ConvertReferenceToObjectName(parentCatalogVersion.Spec.ParentName)
		if parentName == catalogversionName {
			return true
		}

		parentCatalogVersion, err := c.CatalogVersionLookupFunc(ctx, parentName, parentCatalogVersion.ObjectMeta.Namespace)
		if err == nil {
			return c.hasParentCircularDependency(ctx, parentCatalogVersion, catalogversionName)
		}
		return false
	}
}

// Validate NO Child CatalogVersion
// Use ChildCatalogVersionLookupFunc to lookup the child catalogversions with labels {"parentName": catalogversion.ObjectMeta.Name}
func (c *CatalogVersionValidator) ValidateChildCatalogVersion(ctx context.Context, catalogversion model.CatalogVersionState) *ErrorField {
	if c.ChildCatalogVersionLookupFunc == nil {
		return nil
	}
	if found, err := c.ChildCatalogVersionLookupFunc(ctx, catalogversion.ObjectMeta.Name, catalogversion.ObjectMeta.Namespace, string(catalogversion.ObjectMeta.UID)); err != nil || found {
		return &ErrorField{
			FieldPath:       "metadata.name",
			Value:           catalogversion.ObjectMeta.Name,
			DetailedMessage: "CatalogVersion has one or more child catalogversions. Update or Deletion is not allowed",
		}
	}
	return nil
}

func (c *CatalogVersionValidator) ConvertInterfaceToCatalogVersion(ref interface{}) model.CatalogVersionState {
	if ref == nil {
		return model.CatalogVersionState{
			Spec: &model.CatalogVersionSpec{},
		}
	}
	if state, ok := ref.(model.CatalogVersionState); ok {
		if state.Spec == nil {
			state.Spec = &model.CatalogVersionSpec{}
		}
		return state
	} else {
		return model.CatalogVersionState{
			Spec: &model.CatalogVersionSpec{},
		}
	}
}
