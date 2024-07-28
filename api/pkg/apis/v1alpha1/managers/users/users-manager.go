/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package users

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

type UsersManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

type UserState struct {
	Id           string   `json:"id"`
	PasswordHash string   `json:"passwordHash,omitempty"`
	Roles        []string `json:"roles,omitempty"`
}

func (s *UsersManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	stateprovider, err := managers.GetVolatileStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		log.Errorf(" M (Users): failed to get state provider %+v", err)
		return err
	}

	return nil
}
func (t *UsersManager) DeleteUser(ctx context.Context, name string) error {
	ctx, span := observability.StartSpan("Users Manager", ctx, &map[string]string{
		"method": "DeleteUser",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, " M (Users): DeleteUser name %s", name)

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
	})
	if err != nil {
		log.DebugfCtx(ctx, " M (Users) : failed to delete user %s", err)
		return err
	}
	return nil
}

func hash(name string, s string) string {
	h := fnv.New32a()
	h.Write([]byte(name + "." + s + ".salt"))
	return fmt.Sprintf("H%d", h.Sum32())
}

func (t *UsersManager) UpsertUser(ctx context.Context, name string, password string, roles []string) error {
	ctx, span := observability.StartSpan("Users Manager", ctx, &map[string]string{
		"method": "UpsertUser",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, " M (Users): UpsertUser name %s", name)

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: UserState{
				Id:           name,
				PasswordHash: hash(name, password),
				Roles:        roles,
			},
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		log.DebugfCtx(ctx, " M (Users) : failed to upsert user %v", err)
		return err
	}
	return nil
}
func (t *UsersManager) CheckUser(ctx context.Context, name string, password string) ([]string, bool) {
	ctx, span := observability.StartSpan("Users Manager", ctx, &map[string]string{
		"method": "CheckUser",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, " M (Users): CheckUser name %s", name)

	getRequest := states.GetRequest{
		ID: name,
	}
	var user states.StateEntry
	user, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		log.DebugfCtx(ctx, " M (Users) : failed to get user %s states", err)
		return nil, false
	}
	var userState UserState
	bytes, _ := json.Marshal(user.Body)
	err = json.Unmarshal(bytes, &userState)
	if err != nil {
		return nil, false
	}

	if hash(name, password) == userState.PasswordHash {
		log.DebugCtx(ctx, " M (Users) : user authenticated")
		return userState.Roles, true
	}

	log.DebugCtx(ctx, " M (Users) : authentication failed")
	return nil, false
}
