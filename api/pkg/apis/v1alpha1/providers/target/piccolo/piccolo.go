/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package piccolo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

const loggerName = "providers.target.piccolo"
const defaultPiccoloApiServer = "http://0.0.0.0:47099"

var sLog = logger.NewLogger(loggerName)

type PiccoloTargetProviderConfig struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type PiccoloTargetProvider struct {
	Config  PiccoloTargetProviderConfig
	Context *contexts.ManagerContext
}

func PiccoloTargetProviderConfigFromMap(properties map[string]string) (PiccoloTargetProviderConfig, error) {
	ret := PiccoloTargetProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["url"]; ok {
		ret.Url = v
	} else {
		ret.Url = defaultPiccoloApiServer
	}
	return ret, nil
}
func (d *PiccoloTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := PiccoloTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return d.Init(config)
}
func (s *PiccoloTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (d *PiccoloTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Piccolo Target Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Info("  P (Piccolo Target): Init()")

	// convert config to PiccoloTargetProviderConfig type
	piccoloConfig, err := toPiccoloTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (Piccolo Target): expected PiccoloTargetProviderConfig: %+v", err)
		return err
	}

	d.Config = piccoloConfig
	return nil
}

func toPiccoloTargetProviderConfig(config providers.IProviderConfig) (PiccoloTargetProviderConfig, error) {
	ret := PiccoloTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (i *PiccoloTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("Piccolo Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.InfofCtx(ctx, "  P (Piccolo Target): getting artifacts: %s - %s, traceId: %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name, span.SpanContext().TraceID().String())

	ret := make([]model.ComponentSpec, 0)
	for _, component := range references {
		properties := component.Component.Properties
		name := properties["workload.name"].(string)

		req, err := http.NewRequest("GET", i.Config.Url+"/scenario/"+name, nil)
		if err != nil {
			sLog.ErrorCtx(ctx, "  P (Piccolo Target): Unable to make Request")
			return nil, err
		}
		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (Piccolo Target): Unable to get workload %s from piccolo", name)
			return nil, err
		}

		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			// respBody, err := io.ReadAll(resp.Body)
			component := model.ComponentSpec{
				Name:       name,
				Properties: make(map[string]interface{}),
			}
			component.Properties["workload.name"] = string(name)
			ret = append(ret, component)
		case http.StatusNotFound:
			err = errors.New("  P (Piccolo Target): Unable to get workload " + name + " from piccolo")
			return nil, err
		}
	}

	return ret, nil
}

func (i *PiccoloTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Piccolo Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.InfofCtx(ctx, "  P (Piccolo Target): applying artifacts: %s - %s, traceId: %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name, span.SpanContext().TraceID().String())

	injections := &model.ValueInjections{
		InstanceId: deployment.Instance.ObjectMeta.Name,
		SolutionId: deployment.Instance.Spec.Solution,
		TargetId:   deployment.ActiveTarget,
	}

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Piccolo Target): failed to validate components: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	if isDryRun {
		err = nil
		return nil, nil
	}

	ret := step.PrepareResultMap()

	for _, component := range step.Components {
		name := model.ReadPropertyCompat(component.Component.Properties, "workload.name", injections)
		if component.Action == model.ComponentUpdate {
			if name == "" {
				err = errors.New("component doesn't have workload.name property")
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				sLog.ErrorfCtx(ctx, "  P (Piccolo Target): %+v, traceId: %s", err, span.SpanContext().TraceID().String())
				return ret, err
			}
			reqBody := bytes.NewBufferString(name)
			resp, err := http.Post(i.Config.Url+"/scenario", "text/plain", reqBody)
			if err != nil || resp.StatusCode != http.StatusCreated {
				sLog.ErrorCtx(ctx, "  P (Piccolo Target): fail to create resource")
				return ret, err
			}

			defer resp.Body.Close()

			ret[component.Component.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.Updated,
				Message: "",
			}
		} else {
			if name == "" {
				err = errors.New("component doesn't have workload.name property")
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.DeleteFailed,
					Message: err.Error(),
				}
				sLog.ErrorfCtx(ctx, "  P (Piccolo Target): %+v, traceId: %s", err, span.SpanContext().TraceID().String())
				return ret, err
			}
			req, err := http.NewRequest("DELETE", i.Config.Url+"/scenario/"+name, nil)
			if err != nil {
				return ret, err
			}

			client := &http.Client{}
			resp, err := client.Do(req)

			if err == nil && resp.StatusCode == http.StatusOK {
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.Deleted,
					Message: "",
				}
			} else {
				err = errors.New("  P (Piccolo Target): Unable to delete workload " + name + " from piccolo")
				return nil, err
			}
		}
	}
	return ret, nil
}

func (*PiccoloTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{"workload.name"},
			OptionalProperties:    []string{},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
			ChangeDetectionProperties: []model.PropertyDesc{
				{Name: "workload.name", IgnoreCase: false, SkipIfMissing: false},
			},
		},
	}
}
