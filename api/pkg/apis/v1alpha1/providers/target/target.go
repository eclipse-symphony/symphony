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

package target

import (
	"context"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
)

type ITargetProvider interface {
	Init(config providers.IProviderConfig) error
	// apply components to a target
	Apply(ctx context.Context, deployment model.DeploymentSpec) error
	// remove components from a target
	Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error
	// get current component states from a target. The desired state is passed in as a reference
	Get(ctx context.Context, deployment model.DeploymentSpec) ([]model.ComponentSpec, error)
	// the target decides if an update is needed based the the current components and deisred components
	// when a provider re-construct state, it may be unable to re-construct some of the properties
	// in such cases, a provider can choose to ignore some property comparisions
	NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool
	// Provider decides if components should be removed
	NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool
}
