/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/users"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/valyala/fasthttp"
)

var rLog = logger.NewLogger("coa.runtime")

type UsersVendor struct {
	vendors.Vendor
	UsersManager *users.UsersManager
}

func (o *UsersVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Users",
		Producer: "Microsoft",
	}
}

func (e *UsersVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*users.UsersManager); ok {
			e.UsersManager = c
		}
	}
	if e.UsersManager == nil {
		return v1alpha2.NewCOAError(nil, "users manager is not supplied", v1alpha2.MissingConfig)
	}
	if config.Properties != nil && config.Properties["test-users"] == "true" {
		e.UsersManager.UpsertUser(context.Background(), "admin", "", nil)
		e.UsersManager.UpsertUser(context.Background(), "reader", "", nil)
		e.UsersManager.UpsertUser(context.Background(), "developer", "", nil)
		e.UsersManager.UpsertUser(context.Background(), "device-manager", "", nil)
		e.UsersManager.UpsertUser(context.Background(), "operator", "", nil)
	}

	return nil
}

func (o *UsersVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "users"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/auth",
			Version: o.Version,
			Handler: o.onAuth,
		},
	}
}

func (c *UsersVendor) onAuth(request v1alpha2.COARequest) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Users Vendor", request.Context, &map[string]string{
		"method": "onAuth",
	})
	defer span.End()
	log.Infof("V (Users): authenticate user %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	var authRequest AuthRequest
	err := json.Unmarshal(request.Body, &authRequest)
	if err != nil {
		log.Errorf("V (Targets): onAuth failed to unmarshall request body, error: %v traceId: %s", err, span.SpanContext().TraceID().String())
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.Unauthorized,
			Body:  []byte(err.Error()),
		})
	}
	roles, b := c.UsersManager.CheckUser(ctx, authRequest.UserName, authRequest.Password)
	if !b {
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.Unauthorized,
			Body:  []byte("login failed"),
		})
	}

	mySigningKey := []byte("SymphonyKey")
	claims := MyCustomClaims{
		authRequest.UserName,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "symphony",
			Subject:   "symphony",
			ID:        "1",
			Audience:  []string{"*"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, _ := token.SignedString(mySigningKey)

	log.Infof("V (Targets): onAuth succeeded, traceId: %s", span.SpanContext().TraceID().String())
	rolesJSON, _ := json.Marshal(roles)
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        []byte(fmt.Sprintf(`{"accessToken":"%s", "tokenType": "Bearer", "username": "%s", "roles": %s}`, ss, authRequest.UserName, rolesJSON)),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
