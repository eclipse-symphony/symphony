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

package contexts

import (
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	logger "github.com/azure/symphony/coa/pkg/logger"
)

type ManagerContext struct {
	Logger         logger.Logger
	PubsubProvider pubsub.IPubSubProvider
	SiteInfo       v1alpha2.SiteInfo
	VencorContext  *VendorContext
}

func (v *ManagerContext) Init(c *VendorContext, p pubsub.IPubSubProvider) error {
	if c != nil {
		v.Logger = c.Logger
	} else {
		v.Logger = logger.NewLogger("coa.runtime")
	}
	if c == nil {
		v.PubsubProvider = p
	} else {
		v.PubsubProvider = c.PubsubProvider
	}
	if c != nil {
		v.SiteInfo = c.SiteInfo
	}
	if c != nil {
		v.VencorContext = c
	}
	return nil
}

func (v *ManagerContext) Publish(feed string, event v1alpha2.Event) error {
	if v.PubsubProvider != nil {
		return v.PubsubProvider.Publish(feed, event)
	}
	return nil
}

func (v *ManagerContext) Subscribe(feed string, handler v1alpha2.EventHandler) error {
	if v.PubsubProvider != nil {
		return v.PubsubProvider.Subscribe(feed, handler)
	}
	return nil
}

type IWithManagerContext interface {
	SetContext(ctx *ManagerContext)
}
