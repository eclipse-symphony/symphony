/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package remote

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/stretchr/testify/assert"
)

func TestRemoteInitFromMap(t *testing.T) {
	provider := RemoteStageProvider{}
	input := map[string]string{}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
}
func TestRemoteProcess(t *testing.T) {
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	ctx := contexts.ManagerContext{}
	ctx.Init(nil, &pubSubProvider)
	ctx.SiteInfo = v1alpha2.SiteInfo{
		SiteId: "hq",
	}
	provider := RemoteStageProvider{}
	input := map[string]string{}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
	provider.SetContext(&ctx)
	sig := make(chan bool)
	succeededCount := 0
	ctx.Subscribe("remote", func(topic string, event v1alpha2.Event) error {
		var job v1alpha2.JobData
		jData, _ := json.Marshal(event.Body)
		err := json.Unmarshal(jData, &job)
		assert.Nil(t, err)
		assert.Equal(t, "child", event.Metadata["site"])
		assert.Equal(t, "task", event.Metadata["objectType"])
		assert.Equal(t, v1alpha2.JobRun, job.Action)
		succeededCount += 1
		sig <- true
		return nil
	})
	_, _, err = provider.Process(context.Background(), ctx, map[string]interface{}{
		"__site": "child",
	})
	assert.Nil(t, err)
	<-sig
	assert.Equal(t, 1, succeededCount)

	_, _, err = provider.Process(context.Background(), ctx, map[string]interface{}{})
	assert.NotNil(t, err)
	assert.Equal(t, "Bad Request: no site found in inputs", err.Error())
}
