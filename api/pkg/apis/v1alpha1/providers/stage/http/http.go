/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

const (
	loggerName   = "providers.stage.http"
	providerName = "P (Http Stage)"
	httpProvider = "http"
)

var (
	msLock                   sync.Mutex
	sLog                     = logger.NewLogger(loggerName)
	once                     sync.Once
	providerOperationMetrics *metrics.Metrics
)

type HttpStageProviderConfig struct {
	Url                string `json:"url"`
	Method             string `json:"method"`
	SuccessCodes       []int  `json:"successCodes,omitempty"`
	WaitUrl            string `json:"wait.url,omitempty"`
	WaitInterval       int    `json:"wait.interval,omitempty"`
	WaitCount          int    `json:"wait.count,omitempty"`
	WaitStartCodes     []int  `json:"wait.start,omitempty"`
	WaitSuccessCodes   []int  `json:"wait.success,omitempty"`
	WaitFailedCodes    []int  `json:"wait.fail,omitempty"`
	WaitExpression     string `json:"wait.expression,omitempty"`
	WaitExpressionType string `json:"wait.expressionType,omitempty"`
}
type HttpStageProvider struct {
	Config  HttpStageProviderConfig
	Context *contexts.ManagerContext
}

func (m *HttpStageProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	sLog.Debug("  P (Http Stage): initialize")

	mockConfig, err := toHttpStageProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (Http Stage): expected HttpStageProviderConfig: %+v", err)
		return err
	}
	m.Config = mockConfig
	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.Errorf("  P (HTTP Stage): failed to create metrics: %+v", err)
			}
		}
	})
	return err
}
func (s *HttpStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toHttpStageProviderConfig(config providers.IProviderConfig) (HttpStageProviderConfig, error) {
	ret := HttpStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *HttpStageProvider) InitWithMap(properties map[string]string) error {
	config, err := MockStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func MockStageProviderConfigFromMap(properties map[string]string) (HttpStageProviderConfig, error) {
	ret := HttpStageProviderConfig{}
	if v, ok := properties["url"]; ok {
		ret.Url = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "missing required property url", v1alpha2.BadConfig)
	}
	if v, ok := properties["method"]; ok {
		ret.Method = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "missing required property method", v1alpha2.BadConfig)
	}
	if v, ok := properties["successCodes"]; ok {
		codes, err := readIntArray(v)
		if err != nil {
			return ret, err
		}
		ret.SuccessCodes = codes
	}
	if v, ok := properties["wait.success"]; ok {
		codes, err := readIntArray(v)
		if err != nil {
			return ret, err
		}
		ret.WaitSuccessCodes = codes
	}
	if v, ok := properties["wait.start"]; ok {
		codes, err := readIntArray(v)
		if err != nil {
			return ret, err
		}
		ret.WaitStartCodes = codes
	}
	if v, ok := properties["wait.fail"]; ok {
		codes, err := readIntArray(v)
		if err != nil {
			return ret, err
		}
		ret.WaitFailedCodes = codes
	}
	if v, ok := properties["wait.url"]; ok {
		ret.WaitUrl = v
	}
	if v, ok := properties["wait.interval"]; ok {
		interval, err := strconv.Atoi(v)
		if err != nil {
			return ret, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to parse wait interval %v", v), v1alpha2.BadConfig)
		}
		ret.WaitInterval = interval
	}
	if v, ok := properties["wait.count"]; ok {
		count, err := strconv.Atoi(v)
		if err != nil {
			return ret, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to parse wait count %v", v), v1alpha2.BadConfig)
		}
		ret.WaitCount = count
	}
	if v, ok := properties["wait.expression"]; ok {
		ret.WaitExpression = v
	}
	if v, ok := properties["wait.expressionType"]; ok {
		ret.WaitExpressionType = v
	} else {
		ret.WaitExpressionType = "symphony"
	}
	return ret, nil
}
func readIntArray(s string) ([]int, error) {
	var codes []int
	for _, code := range strings.Split(s, ",") {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		intCode, err := strconv.Atoi(code)
		if err != nil {
			return nil, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to parse code %v", code), v1alpha2.BadConfig)
		}
		codes = append(codes, intCode)
	}
	return codes, nil
}
func (i *HttpStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	ctx, span := observability.StartSpan("[Stage] Http provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (Http Stage): start process request")
	functionName := observ_utils.GetFunctionName()
	processTime := time.Now().UTC()
	defer providerOperationMetrics.ProviderOperationLatency(
		processTime,
		httpProvider,
		metrics.ProcessOperation,
		metrics.RunOperationType,
		functionName,
	)

	// Check all config fields for override in inputs
	var configMap map[string]interface{}
	configJson, _ := json.Marshal(i.Config)
	json.Unmarshal(configJson, &configMap)
	for key := range configMap {
		val, found := inputs[key]
		if found {
			configMap[key] = val
		}
	}
	configJson, err = json.Marshal(configMap)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Http Stage): failed to override config with input: %v", err)
		providerOperationMetrics.ProviderOperationErrors(
			httpProvider,
			functionName,
			metrics.ProcessOperation,
			metrics.ValidateOperationType,
			v1alpha2.BadConfig.String(),
		)
		return nil, false, err
	}
	err = json.Unmarshal(configJson, &i.Config)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Http Stage): failed to override config with input: %v", err)
		providerOperationMetrics.ProviderOperationErrors(
			httpProvider,
			functionName,
			metrics.ProcessOperation,
			metrics.ValidateOperationType,
			v1alpha2.BadConfig.String(),
		)
		return nil, false, err
	}

	sLog.InfofCtx(ctx, "  P (Http Stage): %v: %v", i.Config.Method, i.Config.Url)
	webClient := &http.Client{}
	var req *http.Request
	observ_utils.EmitUserAuditsLogs(ctx, "  P (Http Stage): sending request to %v", i.Config.Url)
	req, err = http.NewRequest(fmt.Sprintf("%v", i.Config.Method), fmt.Sprintf("%v", i.Config.Url), nil)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Http Stage): failed to create request: %v", err)
		providerOperationMetrics.ProviderOperationErrors(
			httpProvider,
			functionName,
			metrics.ProcessOperation,
			metrics.RunOperationType,
			v1alpha2.HttpNewRequestFailed.String(),
		)
		return nil, false, err
	}
	for key, input := range inputs {
		if strings.HasPrefix(key, "header.") {
			req.Header.Add(key[7:], fmt.Sprintf("%v", input))
		}
		if key == "body" {
			var jData []byte
			jData, err = json.Marshal(input)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Http Stage): failed to encode json request body: %v", err)
				providerOperationMetrics.ProviderOperationErrors(
					httpProvider,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.HttpNewRequestFailed.String(),
				)
				return nil, false, err
			}
			req.Body = io.NopCloser(bytes.NewBuffer(jData))
			req.Header.Set("Content-Type", "application/json; charset=UTF-8")
			req.ContentLength = int64(len(jData))
		}
	}

	var resp *http.Response
	resp, err = webClient.Do(req)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Http Stage): request failed: %v", err)
		providerOperationMetrics.ProviderOperationErrors(
			httpProvider,
			functionName,
			metrics.ProcessOperation,
			metrics.RunOperationType,
			v1alpha2.HttpSendRequestFailed.String(),
		)
		return nil, false, err
	}
	defer resp.Body.Close()
	outputs := make(map[string]interface{})

	for key, values := range resp.Header {
		outputs[fmt.Sprintf("header.%v", key)] = values
	}

	var data []byte
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Http Stage): failed to read request response: %v", err)
		providerOperationMetrics.ProviderOperationErrors(
			httpProvider,
			functionName,
			metrics.ProcessOperation,
			metrics.RunOperationType,
			v1alpha2.HttpErrorResponse.String(),
		)
		return nil, false, err
	}
	outputs["body"] = string(data) //TODO: probably not so good to assume string
	outputs["status"] = resp.StatusCode

	if i.Config.WaitUrl != "" {
		okToWait := false
		if len(i.Config.WaitStartCodes) > 0 {
			for _, code := range i.Config.WaitStartCodes {
				if code == resp.StatusCode {
					okToWait = true
					break
				}
			}
		}
		if !okToWait {
			providerOperationMetrics.ProviderOperationErrors(
				httpProvider,
				functionName,
				metrics.ProcessOperation,
				metrics.RunOperationType,
				v1alpha2.BadConfig.String(),
			)
			return nil, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("unexpected status code %v", resp.StatusCode), v1alpha2.BadConfig)
		}
		counter := 0
		failed := false
		succeeded := false
		sLog.DebugfCtx(ctx, "  P (Http Stage): WaitCount: %d", i.Config.WaitCount)
		for counter < i.Config.WaitCount || i.Config.WaitCount == 0 {
			sLog.InfofCtx(ctx, "  P (Http Stage): start wait iteration %d", counter)
			var waitReq *http.Request
			waitReq, err = http.NewRequest("GET", i.Config.WaitUrl, nil)
			for key, input := range inputs {
				if strings.HasPrefix(key, "header.") {
					waitReq.Header.Add(key[7:], fmt.Sprintf("%v", input))
				}
			}
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Http Stage): failed to create wait request: %v", err)
				providerOperationMetrics.ProviderOperationErrors(
					httpProvider,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.HttpNewWaitRequestFailed.String(),
				)
				return nil, false, err
			}
			var waitResp *http.Response
			waitResp, err = webClient.Do(waitReq)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Http Stage): wait request failed: %v", err)
				providerOperationMetrics.ProviderOperationErrors(
					httpProvider,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.HttpSendWaitRequestFailed.String(),
				)
				return nil, false, err
			}
			defer waitResp.Body.Close()
			if len(i.Config.WaitFailedCodes) > 0 {
				for _, code := range i.Config.WaitFailedCodes {
					if code == waitResp.StatusCode {
						failed = true
						break
					}
				}
			}
			if len(i.Config.WaitSuccessCodes) > 0 {
				for _, code := range i.Config.WaitSuccessCodes {
					if code == waitResp.StatusCode {
						succeeded = true
						break
					}
				}
			}
			if succeeded {
				var data []byte
				data, err = io.ReadAll(waitResp.Body)
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (Http Stage): failed to read wait request response: %v", err)
					providerOperationMetrics.ProviderOperationErrors(
						httpProvider,
						functionName,
						metrics.ProcessOperation,
						metrics.RunOperationType,
						v1alpha2.HttpErrorWaitResponse.String(),
					)
					succeeded = false
				} else {
					if i.Config.WaitExpression != "" {
						var obj interface{}
						err = json.Unmarshal(data, &obj)
						if err != nil {
							sLog.ErrorfCtx(ctx, "  P (Http Stage): wait response could not be decoded to json: %v", err)
							providerOperationMetrics.ProviderOperationErrors(
								httpProvider,
								functionName,
								metrics.ProcessOperation,
								metrics.RunOperationType,
								v1alpha2.BadConfig.String(),
							)
							succeeded = false
						} else {
							switch i.Config.WaitExpressionType {
							case "jsonpath":
								var result interface{}
								result, err = utils.JsonPathQuery(obj, i.Config.WaitExpression)
								if err != nil {
									sLog.ErrorfCtx(ctx, "  P (Http Stage): failed to evaluate JsonPath: %v", err)
									providerOperationMetrics.ProviderOperationErrors(
										httpProvider,
										functionName,
										metrics.ProcessOperation,
										metrics.RunOperationType,
										v1alpha2.HttpBadWaitExpression.String(),
									)
								}
								succeeded = err == nil
								outputs["waitResult"] = result
							default:
								parser := utils.NewParser(i.Config.WaitExpression)
								var val interface{}
								val, err = parser.Eval(coa_utils.EvaluationContext{
									Value:   obj,
									Context: ctx,
								})
								if err != nil {
									sLog.ErrorfCtx(ctx, "  P (Http Stage): failed to evaluate Symphony expression: %v", err)
									providerOperationMetrics.ProviderOperationErrors(
										httpProvider,
										functionName,
										metrics.ProcessOperation,
										metrics.RunOperationType,
										v1alpha2.HttpBadWaitExpression.String(),
									)
								}
								succeeded = (err == nil && val != "false") // a boolean Symphony expression may evaluate to "false" as a string, indicating the condition is not met
								outputs["waitResult"] = val
							}
						}
					}
					if succeeded {
						outputs["waitBody"] = string(data) //TODO: probably not so good to assume string
					}
				}
			}
			if !failed && !succeeded {
				counter++
				if i.Config.WaitInterval > 0 {
					sLog.DebugCtx(ctx, "  P (Http Stage): sleep for wait interval")
					time.Sleep(time.Duration(i.Config.WaitInterval) * time.Second)
				}
			} else {
				break
			}
		}
		if failed {
			providerOperationMetrics.ProviderOperationErrors(
				httpProvider,
				functionName,
				metrics.ProcessOperation,
				metrics.RunOperationType,
				v1alpha2.HttpBadWaitStatusCode.String(),
			)
			return nil, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to wait for operation %v", resp.StatusCode), v1alpha2.BadConfig)
		}

	} else if len(i.Config.SuccessCodes) > 0 {
		for _, code := range i.Config.SuccessCodes {
			if code == resp.StatusCode {
				return outputs, false, nil
			}
		}
		sLog.ErrorfCtx(ctx, "  P (Http Stage): failed to process request: %d", resp.StatusCode)
		return nil, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("unexpected status code %v", resp.StatusCode), v1alpha2.BadConfig)
	}

	sLog.InfofCtx(ctx, "  P (Http Stage): process request completed with: %d", resp.StatusCode)
	return outputs, false, nil
}
func (*HttpStageProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:        []string{},
			OptionalProperties:        []string{"header.*", "body"},
			RequiredComponentType:     "",
			RequiredMetadata:          []string{},
			OptionalMetadata:          []string{},
			ChangeDetectionProperties: []model.PropertyDesc{},
		},
	}
}
