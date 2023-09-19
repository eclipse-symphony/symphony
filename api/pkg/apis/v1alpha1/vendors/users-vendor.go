/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package vendors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/users"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/azure/symphony/coa/pkg/logger"
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
	log.Debug("V (Users): authenticate user")

	var authRequest AuthRequest
	err := json.Unmarshal(request.Body, &authRequest)
	if err != nil {
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

	rolesJSON, _ := json.Marshal(roles)
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        []byte(fmt.Sprintf(`{"accessToken":"%s", "tokenType": "Bearer", "username": "%s", "roles": %s}`, ss, authRequest.UserName, rolesJSON)),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	span.End()
	return resp
}
