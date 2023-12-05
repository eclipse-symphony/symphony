/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"

	v1alpha2 "github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/valyala/fasthttp"
)

type JWT struct {
	AuthHeader  string                 `json:"authHeader"`
	VerifyKey   string                 `json:"verifyKey"`
	MustHave    []string               `json:"mustHave,omitempty"`
	MustMatch   map[string]interface{} `json:"mustMatch,omitempty"`
	verifyKey   *rsa.PublicKey
	IgnorePaths []string          `json:"ignorePaths,omitempty"`
	Roles       []ClaimRoleMap    `json:"roles,omitempty"`
	EnableRBAC  bool              `json:"enableRBAC,omitempty"`
	Policy      map[string]Policy `json:"policy,omitempty"`
}
type ClaimRoleMap struct {
	Role  string `json:"role"`
	Claim string `json:"claim"`
	Value string `json:"value"`
}
type Policy struct {
	Items map[string]string `json:"items"`
}

func (j JWT) JWT(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if j.IgnorePaths != nil {
			for _, p := range j.IgnorePaths {
				if p == string(ctx.Path()) {
					next(ctx)
					return
				}
			}
		}
		if ctx.IsOptions() {
			next(ctx)
			return
		}
		tokenStr := j.readAuthHeader(ctx)
		if tokenStr == "" {
			ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
		} else {
			_, roles, err := j.validateToken(tokenStr)
			if err != nil {
				ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
			} else {
				if j.EnableRBAC {
					path := string(ctx.Path())
					method := string(ctx.Method())
					for _, role := range roles {
						if v, ok := j.Policy[role]; ok {
							for key, val := range v.Items {
								if key == "*" || strings.HasPrefix(path, key) {
									if val == "*" || strings.Contains(val, method) {
										next(ctx)
										return
									}
								}
							}
						}
					}
					ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
					return
				}
				next(ctx)
			}
		}
	}
}
func (j JWT) readAuthHeader(ctx *fasthttp.RequestCtx) string {
	v := ctx.Request.Header.Peek(j.AuthHeader)
	if v != nil {
		tokenStr := string(v)
		token := strings.Split(tokenStr, "Bearer ")
		if len(token) == 2 {
			return strings.TrimSpace(token[1])
		} else {
			return ""
		}
	}
	return ""
}
func (j *JWT) validateToken(tokenStr string) (map[string]interface{}, []string, error) {
	ret := make(map[string]interface{})
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if j.verifyKey != nil {
				return j.verifyKey, nil
			} else {
				if strings.HasPrefix(j.VerifyKey, "-----BEGIN PUBLIC KEY-----") {
					verifyKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(j.VerifyKey))
					if err != nil {
						return ret, v1alpha2.NewCOAError(nil, "failed to parse public key", v1alpha2.BadConfig)
					}
					j.verifyKey = verifyKey
					return j.verifyKey, nil
				} else {
					return []byte(j.VerifyKey), nil
				}
			}
		},
	)
	if err != nil {
		return ret, nil, err
	}
	if !token.Valid {
		return ret, nil, errors.New("invalid token")
	}
	for k, v := range claims {
		ret[k] = v
	}
	if j.MustHave != nil && len(j.MustHave) > 0 {
		for _, k := range j.MustHave {
			if _, ok := ret[k]; !ok {
				return ret, nil, fmt.Errorf("required claim '%s' is not found", k)
			}
		}
	}
	if j.MustMatch != nil && len(j.MustMatch) > 0 {
		for k, v := range j.MustMatch {
			if hv, ok := ret[k]; ok {
				if hv != v {
					return ret, nil, fmt.Errorf("claim '%s' doesn't have required value", k)
				}
			} else {
				return ret, nil, fmt.Errorf("required claim '%s' is not found", k)
			}
		}
	}
	var roles []string
	if j.EnableRBAC {
		roles = make([]string, 0)
		for _, m := range j.Roles {
			if v, ok := ret[m.Claim]; ok {
				if m.Value == "*" || v == m.Value {
					roles = append(roles, m.Role)
				}
			}
		}

	}
	return ret, roles, nil
}
