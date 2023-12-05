/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/valyala/fasthttp"
)

type Trail struct {
	PubSubProvider pubsub.IPubSubProvider
}

func (j Trail) Trail(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		next(ctx)
	}
}
func (j *Trail) SetPubSubProvider(provider pubsub.IPubSubProvider) {
	j.PubSubProvider = provider
}
