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

// Check Catalog existence
var CatalogLookupFunc ObjectLookupFunc
var CatalogContainerLookupFunc ObjectLookupFunc
var ChildCatalogLookupFunc LinkedObjectLookupFunc

// Validate Catalog creation or update
// 1. Schema is valid
// 2. Parent catalog exists
// 3. Catalog name and rootResource is valid. And rootResource is immutable
// TODO: 4. Update won't form a cycle in the parent-child relationship
func ValidateCreateOrUpdateCatalog(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := ConvertInterfaceToCatalog(newRef)
	old := ConvertInterfaceToCatalog(oldRef)

	errorFields := []ErrorField{}
	if err := ValidateSchema(ctx, new); err != nil {
		errorFields = append(errorFields, *err)
	}
	if new.Spec.ParentName != "" && (oldRef == nil || new.Spec.ParentName != old.Spec.ParentName) {
		if err := ValidateParentCatalog(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}

	if oldRef == nil {
		// validate create specific fields
		if err := ValidateObjectName(new.ObjectMeta.Name, new.Spec.RootResource); err != nil {
			errorFields = append(errorFields, *err)
		}
		// validate rootResource
		if err := ValidateRootResource(ctx, new.ObjectMeta, new.Spec.RootResource, CatalogContainerLookupFunc); err != nil {
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

// Validate Catalog deletion
// 1. Catalog has no child catalogs
func ValidateDeleteCatalog(ctx context.Context, catalog interface{}) []ErrorField {
	new := ConvertInterfaceToCatalog(catalog)
	// validate child catalogs
	errorFields := []ErrorField{}
	if err := ValidateChildCatalog(ctx, new); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

// Validate Schema is valid
func ValidateSchema(ctx context.Context, new model.CatalogState) *ErrorField {
	if CatalogLookupFunc == nil {
		return nil
	}
	if schemaName, ok := new.Spec.Metadata["schema"]; ok {
		// 1). Lookup catalog object with schema name
		schemaName = ConvertReferenceToObjectName(schemaName)
		lookupRes, err := CatalogLookupFunc(ctx, schemaName, new.ObjectMeta.Namespace)

		if err != nil {
			return &ErrorField{
				FieldPath:       "spec.metadata.schema",
				Value:           schemaName,
				DetailedMessage: "could not find the required schema",
			}
		}
		marshalResult, _ := json.Marshal(lookupRes)
		var catalog model.CatalogState
		err = json.Unmarshal(marshalResult, &catalog)
		if err != nil {
			return &ErrorField{
				FieldPath:       "spec.metadata.schema",
				Value:           schemaName,
				DetailedMessage: "schema is not a valid catalog object",
			}
		}
		if spec, ok := catalog.Spec.Properties["spec"]; ok {
			// 2). Extract Schema object from the catalog object
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

			// 3). Validate the schema on the catalog which is being created/updated
			result, err := schemaObj.CheckProperties(new.Spec.Properties, nil)
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

// Validate Parent Catalog exists if provided
// Use CatalogLookupFunc to lookup the catalog with parentName
func ValidateParentCatalog(ctx context.Context, catalog model.CatalogState) *ErrorField {
	if CatalogLookupFunc == nil {
		return nil
	}
	parentCatalogName := ConvertReferenceToObjectName(catalog.Spec.ParentName)
	_, err := CatalogLookupFunc(ctx, parentCatalogName, catalog.ObjectMeta.Namespace)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.ParentName",
			Value:           parentCatalogName,
			DetailedMessage: "parent catalog not found",
		}
	}
	return nil
}

// Validate NO Child Catalog
// Use ChildCatalogLookupFunc to lookup the child catalogs with labels {"parentName": catalog.ObjectMeta.Name}
func ValidateChildCatalog(ctx context.Context, catalog model.CatalogState) *ErrorField {
	if ChildCatalogLookupFunc == nil {
		return nil
	}
	if found, err := ChildCatalogLookupFunc(ctx, catalog.ObjectMeta.Name, catalog.ObjectMeta.Namespace); err != nil || found {
		return &ErrorField{
			FieldPath:       "metadata.name",
			Value:           catalog.ObjectMeta.Name,
			DetailedMessage: "Catalog has one or more child catalogs. Update or Deletion is not allowed",
		}
	}
	return nil
}

func ConvertInterfaceToCatalog(ref interface{}) model.CatalogState {
	if ref == nil {
		return model.CatalogState{
			Spec: &model.CatalogSpec{},
		}
	}
	if state, ok := ref.(model.CatalogState); ok {
		if state.Spec == nil {
			state.Spec = &model.CatalogSpec{}
		}
		return state
	} else {
		return model.CatalogState{
			Spec: &model.CatalogSpec{},
		}
	}
}
