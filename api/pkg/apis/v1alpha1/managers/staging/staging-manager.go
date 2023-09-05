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

package staging

import (
	"context"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/queue"
)

type StagingManager struct {
	managers.Manager
	QueueProvider queue.IQueueProvider
}

func (s *StagingManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	queueProvider, err := managers.GetQueueProvider(config, providers)
	if err == nil {
		s.QueueProvider = queueProvider
	} else {
		return err
	}
	return nil
}
func (s *StagingManager) Enabled() bool {
	return s.Config.Properties["poll.enabled"] == "true"
}
func (s *StagingManager) Poll() []error {
	return nil
}
func (s *StagingManager) Reconcil() []error {
	return nil
}

func (s *StagingManager) HandleJobEvent(ctx context.Context, event v1alpha2.Event) error {
	var job v1alpha2.JobData
	var jok bool
	if job, jok = event.Body.(v1alpha2.JobData); !jok {
		return v1alpha2.NewCOAError(nil, "event body is not a job", v1alpha2.BadRequest)
	}
	return s.QueueProvider.Enqueue(event.Metadata["site"], job)
}
func (s *StagingManager) GetABatchForSite(site string) ([]v1alpha2.JobData, error) {
	//TODO: this should return a group of jobs as optimization
	if s.QueueProvider.Size(site) == 0 {
		return nil, nil
	}
	stackElement, err := s.QueueProvider.Dequeue(site)
	if err != nil {
		return nil, err
	}
	if job, ok := stackElement.(v1alpha2.JobData); ok {
		return []v1alpha2.JobData{
			job,
		}, nil
	} else {
		s.QueueProvider.Enqueue(site, stackElement)
	}
	return nil, nil
}
