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

package users

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/azure/symphony/coa/pkg/logger"
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
	stateprovider, err := managers.GetStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}

	return nil
}
func (t *UsersManager) DeleteUser(ctx context.Context, name string) error {
	return t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
	})
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
	log.Debug(" M (Users) : upsert user")
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
	_, err := t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		log.Debugf(" M (Users) : failed to upsert user - %s", err)
		return err
	}
	observ_utils.CloseSpanWithError(span, nil)
	return nil
}
func (t *UsersManager) CheckUser(ctx context.Context, name string, password string) ([]string, bool) {
	ctx, span := observability.StartSpan("Users Manager", ctx, &map[string]string{
		"method": "CheckUser",
	})
	log.Debug(" M (Users) : check user")
	getRequest := states.GetRequest{
		ID: name,
	}
	user, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		log.Debugf(" M (Users) : failed to read user - %s", err)
		return nil, false
	}

	if v, ok := user.Body.(UserState); ok {
		if hash(name, password) == v.PasswordHash {
			observ_utils.CloseSpanWithError(span, nil)
			log.Debug(" M (Users) : user authenticated")
			return v.Roles, true
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	log.Debug(" M (Users) : authentication failed")
	return nil, false
}
