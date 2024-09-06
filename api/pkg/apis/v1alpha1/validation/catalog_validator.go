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

type CatalogValidator struct {
	// Check Catalog existence
	CatalogLookupFunc          ObjectLookupFunc
	CatalogContainerLookupFunc ObjectLookupFunc
	ChildCatalogLookupFunc     LinkedObjectLookupFunc
}

func NewCatalogValidator(catalogLookupFunc ObjectLookupFunc, catalogContainerLookupFunc ObjectLookupFunc, childCatalogLookupFunc LinkedObjectLookupFunc) CatalogValidator {
	return CatalogValidator{
		CatalogLookupFunc:          catalogLookupFunc,
		CatalogContainerLookupFunc: catalogContainerLookupFunc,
		ChildCatalogLookupFunc:     childCatalogLookupFunc,
	}
}

// Validate Catalog creation or update
// 1. Schema is valid
// 2. Parent catalog exists
// 3. Catalog name and rootResource is valid. And rootResource is immutable
// TODO: 4. Update won't form a cycle in the parent-child relationship
func (c *CatalogValidator) ValidateCreateOrUpdate(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := c.ConvertInterfaceToCatalog(newRef)
	old := c.ConvertInterfaceToCatalog(oldRef)

	errorFields := []ErrorField{}
	if err := c.ValidateSchema(ctx, new); err != nil {
		errorFields = append(errorFields, *err)
	}
	if new.Spec.ParentName != "" && (oldRef == nil || new.Spec.ParentName != old.Spec.ParentName) {
		if err := c.ValidateParentCatalog(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}

	if oldRef == nil {
		// validate create specific fields
		if err := ValidateObjectName(new.ObjectMeta.Name, new.Spec.RootResource); err != nil {
			errorFields = append(errorFields, *err)
		}
		// validate rootResource
		if err := ValidateRootResource(ctx, new.ObjectMeta, new.Spec.RootResource, c.CatalogContainerLookupFunc); err != nil {
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
func (c *CatalogValidator) ValidateDelete(ctx context.Context, catalog interface{}) []ErrorField {
	new := c.ConvertInterfaceToCatalog(catalog)
	// validate child catalogs
	errorFields := []ErrorField{}
	if err := c.ValidateChildCatalog(ctx, new); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

// Validate Schema is valid
func (c *CatalogValidator) ValidateSchema(ctx context.Context, new model.CatalogState) *ErrorField {
	if c.CatalogLookupFunc == nil {
		return nil
	}
	if schemaName, ok := new.Spec.Metadata["schema"]; ok {
		// 1). Lookup catalog object with schema name
		schemaName = ConvertReferenceToObjectName(schemaName)
		lookupRes, err := c.CatalogLookupFunc(ctx, schemaName, new.ObjectMeta.Namespace)

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
func (c *CatalogValidator) ValidateParentCatalog(ctx context.Context, catalog model.CatalogState) *ErrorField {
	if c.CatalogLookupFunc == nil {
		return nil
	}
	parentCatalogName := ConvertReferenceToObjectName(catalog.Spec.ParentName)
	parentCatalog, err := c.CatalogLookupFunc(ctx, parentCatalogName, catalog.ObjectMeta.Namespace)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.ParentName",
			Value:           parentCatalogName,
			DetailedMessage: "parent catalog not found",
		}
	}

	// if parent exists, need to check if upsert this catalog will introduce circular parent issue.
	catalogName := catalog.ObjectMeta.Name
	if c.hasParentCircularDependency(ctx, parentCatalog, catalogName) {
		return &ErrorField{
			FieldPath:       "spec.ParentName",
			Value:           parentCatalogName,
			DetailedMessage: "parent catalog has circular dependency",
		}
	}
	return nil
}

func (c *CatalogValidator) hasParentCircularDependency(ctx context.Context, parentRaw interface{}, catalogName string) bool {
	jsonData, err := json.Marshal(parentRaw)
	if err != nil {
		return false
	}

	var parentCatalog model.CatalogState
	err = json.Unmarshal(jsonData, &parentCatalog)
	if err != nil {
		return false
	}

	if parentCatalog.Spec.ParentName == "" {
		return false
	} else {
		parentName := ConvertReferenceToObjectName(parentCatalog.Spec.ParentName)
		if parentName == catalogName {
			return true
		}

		parentCatalog, err := c.CatalogLookupFunc(ctx, parentName, parentCatalog.ObjectMeta.Namespace)
		if err == nil {
			return c.hasParentCircularDependency(ctx, parentCatalog, catalogName)
		}
		return false
	}
}

// Validate NO Child Catalog
// Use ChildCatalogLookupFunc to lookup the child catalogs with labels {"parentName": catalog.ObjectMeta.Name}
func (c *CatalogValidator) ValidateChildCatalog(ctx context.Context, catalog model.CatalogState) *ErrorField {
	if c.ChildCatalogLookupFunc == nil {
		return nil
	}
	if found, err := c.ChildCatalogLookupFunc(ctx, catalog.ObjectMeta.Name, catalog.ObjectMeta.Namespace); err != nil || found {
		return &ErrorField{
			FieldPath:       "metadata.name",
			Value:           catalog.ObjectMeta.Name,
			DetailedMessage: "Catalog has one or more child catalogs. Update or Deletion is not allowed",
		}
	}
	return nil
}

func (c *CatalogValidator) ConvertInterfaceToCatalog(ref interface{}) model.CatalogState {
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
