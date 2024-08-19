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

func ValidateDeleteCatalog(ctx context.Context, catalog interface{}) []ErrorField {
	new := ConvertInterfaceToCatalog(catalog)
	// validate child catalogs
	errorFields := []ErrorField{}
	if err := ValidateChildCatalog(ctx, new); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

func ValidateSchema(ctx context.Context, new model.CatalogState) *ErrorField {
	if CatalogLookupFunc == nil {
		return nil
	}
	if schemaName, ok := new.Spec.Metadata["schema"]; ok {
		schemaName = ConvertReferenceToObjectName(schemaName)
		//cataloglog.Info("Find schema name", "name", schemaName)
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
			result, err := schemaObj.CheckProperties(new.Spec.Properties, nil)
			if err != nil {
				//cataloglog.Error(err, "Validating failed.")
				return &ErrorField{
					FieldPath:       "spec.metadata.schema",
					Value:           schemaName,
					DetailedMessage: "unable to determine the validity of the schema",
				}
			}
			if !result.Valid {
				//cataloglog.Error(err, "Validating failed.")
				return &ErrorField{
					FieldPath:       "spec.Properties",
					Value:           "(hidden)",
					DetailedMessage: "invalid schema result: " + result.ToErrorMessages(),
				}
			}
		}
		//cataloglog.Info("Validation finished.", "name", r.Name)
	}
	return nil
}

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
