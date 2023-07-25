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

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/stage"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/azure/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

type StageVendor struct {
	vendors.Vendor
	StageManager *stage.StageManager
}

func (s *StageVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  s.Vendor.Version,
		Name:     "Stage",
		Producer: "Microsoft",
	}
}

func (o *StageVendor) GetEndpoints() []v1alpha2.Endpoint {
	return []v1alpha2.Endpoint{}
}

func (s *StageVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := s.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range s.Managers {
		if c, ok := m.(*stage.StageManager); ok {
			s.StageManager = c
		}
	}
	if s.StageManager == nil {
		return v1alpha2.NewCOAError(nil, "stage manager is not supplied", v1alpha2.MissingConfig)
	}
	s.Vendor.Context.Subscribe("activation", func(topic string, event v1alpha2.Event) error {
		evt, err := s.StageManager.HandleActivationEvent(context.Background(), event)
		if err != nil {
			return err
		}
		if evt != nil {
			s.Vendor.Context.Publish("trigger", v1alpha2.Event{
				Body: *evt,
			})
		}
		return nil
	})
	s.Vendor.Context.Subscribe("trigger", func(topic string, event v1alpha2.Event) error {
		status, err := s.StageManager.HandleTriggerEvent(context.Background(), event)
		if err != nil {
			return err
		}
		if status != nil {
			s.Vendor.Context.Publish("trigger", v1alpha2.Event{
				Body: *status,
			})
		}
		return nil
	})
	return nil
}
