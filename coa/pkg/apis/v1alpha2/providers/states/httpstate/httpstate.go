/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package httpstate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	states "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

type HttpStateProviderConfig struct {
	Name              string `json:"name"`
	Url               string `json:"url"`
	PostAsArray       bool   `json:"postAsArray,omitempty"`
	PostNameInPath    bool   `json:"postNameInPath,omitempty"`
	PostBodyKeyName   string `json:"postBodyKeyName,omitempty"`
	PostBodyValueName string `json:"postBodyValueName,omitempty"`
	NotFoundAs204     bool   `json:"notFoundAs204,omitempty"`
}

type HttpStateProvider struct {
	Config  HttpStateProviderConfig
	Data    map[string]interface{}
	Context *contexts.ManagerContext
}

func HttpStateProviderConfigFromMap(properties map[string]string) (HttpStateProviderConfig, error) {
	ret := HttpStateProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	if v, ok := properties["postBodyKeyName"]; ok {
		ret.PostBodyKeyName = utils.ParseProperty(v)
	}
	if v, ok := properties["postBodyValueName"]; ok {
		ret.PostBodyValueName = utils.ParseProperty(v)
	}
	if v, ok := properties["postAsArray"]; ok {
		val := utils.ParseProperty(v)
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'postAsArray' setting of Http state provider", v1alpha2.BadConfig)
			}
			ret.PostAsArray = bVal
		}
	}
	if v, ok := properties["postNameInPath"]; ok {
		val := utils.ParseProperty(v)
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'postNameInPath' setting of Http state provider", v1alpha2.BadConfig)
			}
			ret.PostNameInPath = bVal
		}
	} else {
		ret.PostNameInPath = true
	}
	if v, ok := properties["notFoundAs204"]; ok {
		val := utils.ParseProperty(v)
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'notFoundAs204' setting of Http state provider", v1alpha2.BadConfig)
			}
			ret.NotFoundAs204 = bVal
		}
	}
	if v, ok := properties["url"]; ok {
		ret.Url = utils.ParseProperty(v)
	} else {
		return ret, v1alpha2.NewCOAError(nil, "Http sate provider url is not set", v1alpha2.BadConfig)
	}
	return ret, nil
}

func (s *HttpStateProvider) ID() string {
	return s.Config.Name
}

func (s *HttpStateProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *HttpStateProvider) InitWithMap(properties map[string]string) error {
	config, err := HttpStateProviderConfigFromMap(properties)
	if err != nil {
		sLog.Errorf("  P (Http State): failed to parse provider config from map %+v", err)
		return err
	}
	return i.Init(config)
}

func (s *HttpStateProvider) Init(config providers.IProviderConfig) error {
	// parameter checks
	stateConfig, err := toHttpStateProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (Http State): failed to parse provider config %+v", err)
		return errors.New("expected HttpStateProviderConfig")
	}
	s.Config = stateConfig
	if s.Config.Url == "" {
		return v1alpha2.NewCOAError(nil, "Http state provider url is not set", v1alpha2.BadConfig)
	}
	s.Data = make(map[string]interface{}, 0)
	return nil
}

func (s *HttpStateProvider) Upsert(ctx context.Context, entry states.UpsertRequest) (string, error) {
	_, span := observability.StartSpan("Http State Provider", ctx, &map[string]string{
		"method": "Upsert",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	sLog.Infof("  P (Http State): upsert states %s, traceId: %s", entry.Value.ID, span.SpanContext().TraceID().String())

	client := &http.Client{}
	rUrl := s.Config.Url
	if entry.Value.ID == "" {
		err = v1alpha2.NewCOAError(nil, "found invalid entry ID", v1alpha2.InternalError)
		sLog.Errorf("  P (Http State): upsert failed %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return "", err
	}
	if s.Config.PostNameInPath {
		rUrl, err = url.JoinPath(s.Config.Url, entry.Value.ID)
	}
	if err != nil {
		sLog.Errorf("  P (Http State): failed to form %s request path: %+v, traceId: %s", entry.Value.ID, err, span.SpanContext().TraceID().String())
		return "", err
	}
	obj := entry.Value.Body
	if s.Config.PostBodyKeyName != "" && s.Config.PostBodyValueName != "" {
		obj = map[string]interface{}{
			s.Config.PostBodyKeyName:   entry.Value.ID,
			s.Config.PostBodyValueName: obj,
		}
	}
	if s.Config.PostAsArray {
		obj = []interface{}{obj}
	}
	jData, _ := json.Marshal(obj)
	req, err := http.NewRequest("POST", rUrl, bytes.NewBuffer(jData))
	if err != nil {
		sLog.Errorf("  P (Http State): failed to create a Post request: %+v, traceId: %s", entry.Value.ID, err, span.SpanContext().TraceID().String())
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		sLog.Errorf("  P (Http State): failed to get response from upserting %s: %+v, traceId: %s", entry.Value.ID, err, span.SpanContext().TraceID().String())
		return "", err
	}
	if resp.StatusCode >= 300 {
		sLog.Errorf("  P (Http State): failed to get correct state code: %+v, status code %d, traceId: %s", entry.Value.ID, resp.StatusCode, err, span.SpanContext().TraceID().String())
		return "", fmt.Errorf("failed to invoke HTTP state store: [%d]", resp.StatusCode)
	}
	return entry.Value.ID, nil
}

func (s *HttpStateProvider) List(ctx context.Context, request states.ListRequest) ([]states.StateEntry, string, error) {
	return nil, "", v1alpha2.NewCOAError(nil, "Http state store list is not implemented", v1alpha2.NotImplemented)
}

func (s *HttpStateProvider) Delete(ctx context.Context, request states.DeleteRequest) error {
	_, span := observability.StartSpan("Http State Provider", ctx, &map[string]string{
		"method": "Delete",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	sLog.Infof("  P (Http State): list states, traceId: %s", span.SpanContext().TraceID().String())

	client := &http.Client{}
	if request.ID == "" {
		err := v1alpha2.NewCOAError(nil, "found invalid request ID", v1alpha2.InternalError)
		sLog.Errorf("  P (Http State): failed to list states: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return err
	}
	rUrl, err := url.JoinPath(s.Config.Url, request.ID)
	if err != nil {
		sLog.Errorf("  P (Http State): failed to form %s request path: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return err
	}
	req, err := http.NewRequest("DELETE", rUrl, nil)
	if err != nil {
		sLog.Errorf("  P (Http State): failed to create a Delete request: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		sLog.Errorf("  P (Http State): failed to get response from upserting %s: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return err
	}
	if resp.StatusCode >= 300 {
		sLog.Errorf("  P (Http State): failed to get correct state code: %+v, status code %d, traceId: %s", request.ID, resp.StatusCode, err, span.SpanContext().TraceID().String())
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to delete from HTTP state store: [%d]", resp.StatusCode), v1alpha2.InternalError)
	}
	return nil
}

func (s *HttpStateProvider) Get(ctx context.Context, request states.GetRequest) (states.StateEntry, error) {
	_, span := observability.StartSpan("Http State Provider", ctx, &map[string]string{
		"method": "Delete",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	sLog.Infof("  P (Http State): get states %s, traceId: %s", request.ID, span.SpanContext().TraceID().String())

	client := &http.Client{}
	if request.ID == "" {
		err := v1alpha2.NewCOAError(nil, "found invalid request ID", v1alpha2.InternalError)
		sLog.Errorf("  P (Http State): failed to get states: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return states.StateEntry{}, err
	}
	rUrl, err := url.JoinPath(s.Config.Url, request.ID)
	if err != nil {
		sLog.Errorf("  P (Http State): failed to create a Get request: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return states.StateEntry{}, err
	}
	req, err := http.NewRequest("GET", rUrl, nil)
	if err != nil {
		sLog.Errorf("  P (Http State): request creation failed: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return states.StateEntry{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		sLog.Errorf("  P (Http State): failed to get response from getting %s: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return states.StateEntry{}, err
	}
	if resp.StatusCode == 204 && s.Config.NotFoundAs204 {
		sLog.Infof("  P (Http State): cannot find %s state: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return states.StateEntry{}, v1alpha2.NewCOAError(nil, "not found", v1alpha2.NotFound)
	}
	if resp.StatusCode >= 300 {
		if resp.StatusCode == 404 {
			sLog.Infof("  P (Http State): received 404 status code: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
			return states.StateEntry{}, v1alpha2.NewCOAError(nil, "not found", v1alpha2.NotFound)
		} else {
			sLog.Errorf("  P (Http State): failed to get correct state code: %+v, status code %d, traceId: %s", request.ID, resp.StatusCode, err, span.SpanContext().TraceID().String())
			return states.StateEntry{}, v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to invoke HTTP state store: [%d]", resp.StatusCode), v1alpha2.InternalError)
		}

	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		sLog.Errorf("  P (Http State): failed to read request body: %+v, traceId: %s", resp.StatusCode, err, span.SpanContext().TraceID().String())
		return states.StateEntry{}, err
	}
	var obj interface{}
	err = json.Unmarshal(bodyBytes, &obj)
	if err != nil {
		sLog.Errorf("  P (Http State): failed to unmarshall response body: %+v, traceId: %s", resp.StatusCode, err, span.SpanContext().TraceID().String())
		return states.StateEntry{}, err
	}
	return states.StateEntry{
		ID:   request.ID,
		Body: obj,
	}, nil
}

func toHttpStateProviderConfig(config providers.IProviderConfig) (HttpStateProviderConfig, error) {
	ret := HttpStateProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	//ret.Name = providers.LoadEnv(ret.Name)
	return ret, err
}

func (a *HttpStateProvider) Clone(config providers.IProviderConfig) (providers.IProvider, error) {
	ret := &HttpStateProvider{}
	if config == nil {
		err := ret.Init(a.Config)
		if err != nil {
			return nil, err
		}
	} else {
		err := ret.Init(config)
		if err != nil {
			return nil, err
		}
	}
	if a.Context != nil {
		ret.Context = a.Context
	}
	return ret, nil
}
