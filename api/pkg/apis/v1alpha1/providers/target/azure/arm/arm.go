/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package arm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/contexts"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var (
	sLog                     = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
)

const (
	arm          = "arm"
	providerName = "P (ARM Target)"
	loggerName   = "providers.target.azure.arm"
)

type (
	ArmTargetProviderConfig struct {
		SubscriptionId string `json:"subscriptionId"`
	}
	ArmTargetProvider struct {
		Config              ArmTargetProviderConfig
		Context             *contexts.ManagerContext
		ResourceGroupClient *armresources.ResourceGroupsClient
		DeploymentsClient   *armresources.DeploymentsClient
	}
	UrlOrJson struct {
		URL  *url.URL    `json:"url,omitempty"`
		JSON interface{} `json:"json,omitempty"`
	}
	ArmDeployment struct {
		ResourceGroup string    `json:"resourceGroup"`
		Location      string    `json:"location"`
		Template      UrlOrJson `json:"template"`
		Parameters    UrlOrJson `json:"parameters,omitempty"`
	}
)

func (u *UrlOrJson) IsEmpty() bool {
	return u.URL == nil && u.JSON == nil
}
func (u *UrlOrJson) GetJson() (map[string]interface{}, error) {
	if u.URL != nil {
		resp, err := http.Get(u.URL.String())
		if err != nil {
			return nil, fmt.Errorf("failed to download JSON from URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to download JSON from URL, status: %s", resp.Status)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("failed to parse JSON from URL: %w", err)
		}

		return data, nil
	}

	if u.JSON != nil {
		switch v := u.JSON.(type) {
		case string:
			// Parse JSON if it's a string
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(v), &data); err != nil {
				return nil, fmt.Errorf("failed to parse inline JSON string: %w", err)
			}
			return data, nil
		case map[string]interface{}:
			return v, nil
		default:
			return nil, fmt.Errorf("unsupported JSON type: %T", v)
		}
	}

	return nil, errors.New("both URL and JSON are empty")
}

func getArmDeploymentFromComponent(component model.ComponentSpec) (*ArmDeployment, error) {
	ret := ArmDeployment{}
	data, err := json.Marshal(component.Properties)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return validateDeployment(&ret)
}
func validateDeployment(deployment *ArmDeployment) (*ArmDeployment, error) {
	if deployment.ResourceGroup == "" {
		return nil, errors.New("resourceGroup is required")
	}
	if deployment.Location == "" {
		return nil, errors.New("location is required")
	}
	if deployment.Template.URL == nil && deployment.Template.JSON == "" {
		return nil, errors.New("template is required")
	}
	return deployment, nil
}

func ArmTargetProviderConfigFromMap(properties map[string]string) (ArmTargetProviderConfig, error) {
	ret := ArmTargetProviderConfig{}
	if v, ok := properties["subscriptionId"]; ok {
		ret.SubscriptionId = v
	} else {
		return ret, errors.New("subscriptionId is required")
	}
	return ret, nil
}

func (r *ArmTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := ArmTargetProviderConfigFromMap(properties)
	if err != nil {
		sLog.Errorf("  P (ARM Target): expected ArmTargetProviderConfig: %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to init", providerName), v1alpha2.InitFailed)
	}

	return r.Init(config)
}

func (s *ArmTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (r *ArmTargetProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan(
		"ARM Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfoCtx(ctx, "  P (ARM Target): Init()")

	updateConfig, err := toARMTargetProviderConfig(config)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (ARM Target): expected ArmTargetProviderConfig - %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to convert to ArmTargetProviderConfig", providerName), v1alpha2.InitFailed)
		return err
	}
	r.Config = updateConfig

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	resourcesClientFactory, err := armresources.NewClientFactory(r.Config.SubscriptionId, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	r.ResourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()
	r.DeploymentsClient = resourcesClientFactory.NewDeploymentsClient()

	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.ErrorCtx(ctx, err)
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to init metrics", providerName), v1alpha2.InitFailed)
			}
		}
	})
	return err
}

func (r *ArmTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan(
		"ARM Target Provider",
		ctx, &map[string]string{
			"method": "Get",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (ARM Target): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	ret := make([]model.ComponentSpec, 0)

	return ret, nil
}
func (r *ArmTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan(
		"ARM Target Provider",
		ctx,
		&map[string]string{
			"method": "Get",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (ARM Target): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	ret := make([]model.ComponentSpec, 0)
	for _, component := range references {
		deploymentProp, err := getArmDeploymentFromComponent(component.Component)
		if err != nil {
			return ret, err
		}
		_, err = r.analyzeDeployment(ctx, deploymentName, deploymentProp)
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}
func (r *ArmTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan(
		"ARM Target Provider",
		ctx,
		&map[string]string{
			"method": "Apply",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (ARM Target):  applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	functionName := utils.GetFunctionName()
	startTime := time.Now().UTC()
	defer providerOperationMetrics.ProviderOperationLatency(
		startTime,
		arm,
		metrics.ApplyOperation,
		metrics.ApplyOperationType,
		functionName,
	)
	components := step.GetComponents()
	err = r.GetValidationRule(ctx).Validate(components)
	if err != nil {
		providerOperationMetrics.ProviderOperationErrors(
			arm,
			functionName,
			metrics.ValidateRuleOperation,
			metrics.ApplyOperationType,
			v1alpha2.ValidateFailed.String(),
		)

		sLog.ErrorfCtx(ctx, "  P (ARM Target): failed to validate components, error: %v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: the rule validation failed", providerName), v1alpha2.ValidateFailed)
		return nil, err
	}
	if isDryRun {
		sLog.DebugfCtx(ctx, "  P (ARM Target): dryRun is enabled,, skipping apply")
		return nil, nil
	}

	ret := step.PrepareResultMap()

	for _, component := range step.Components {
		deploymentProp, err := getArmDeploymentFromComponent(component.Component)
		deploymentName := deployment.Instance.ObjectMeta.Name + "_" + component.Component.Name
		if component.Action == model.ComponentUpdate {
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (ARM Target): failed to get ARM deployment: %+v", err)
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get ARM deployment", providerName), v1alpha2.GetARMDeploymentPropertyFailed)
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				providerOperationMetrics.ProviderOperationErrors(
					arm,
					functionName,
					metrics.ARMDeploymentPropertyOperation,
					metrics.ApplyOperationType,
					v1alpha2.GetARMDeploymentPropertyFailed.String(),
				)
				return ret, err
			}
			utils.EmitUserAuditsLogs(ctx, "  P (ARM Target): Creating ARM deployment: %s", deploymentName)
			_, err := r.ensureResourceGroup(ctx, r.Config.SubscriptionId, deploymentProp.Location, deploymentProp.ResourceGroup)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (ARM Target): failed to ensure resource group: %+v", err)
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to ensure resource group", providerName), v1alpha2.EnsureARMResourceGroupFailed)
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				providerOperationMetrics.ProviderOperationErrors(
					arm,
					functionName,
					metrics.ARMResourceGroupOperation,
					metrics.ApplyOperationType,
					v1alpha2.EnsureARMResourceGroupFailed.String(),
				)
				return ret, err
			}
			_, err = r.createDeployment(ctx, deploymentName, deploymentProp, false)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (ARM Target): failed to create ARM deployment: %+v", err)
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to create ARM deployment", providerName), v1alpha2.CreateARMDeploymentFailed)
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				providerOperationMetrics.ProviderOperationErrors(
					arm,
					functionName,
					metrics.ARMCreateDeploymentOperation,
					metrics.ApplyOperationType,
					v1alpha2.CreateARMDeploymentFailed.String(),
				)
				return ret, err
			}

			sLog.InfofCtx(ctx, "  P (ARM Target): created ARM deployment successfully: %s", component.Component.Name)
			ret[component.Component.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.Updated,
				Message: fmt.Sprintf("No error. %s has been updated", component.Component.Name),
			}
		} else {
			utils.EmitUserAuditsLogs(ctx, "  P (ARM Target): Cleaning up ARM deployment: %s", deploymentName)
			err = r.cleanUpDeployment(ctx, deploymentName, deploymentProp)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (ARM Target): failed to clean up ARM deployment: %+v", err)
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to clean up ARM deployment", providerName), v1alpha2.CleanUpARMDeploymentFailed)
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				providerOperationMetrics.ProviderOperationErrors(
					arm,
					functionName,
					metrics.ARMCleanUpDeploymentOperation,
					metrics.ApplyOperationType,
					v1alpha2.CleanUpARMDeploymentFailed.String(),
				)
				return ret, err
			}
			sLog.InfofCtx(ctx, "  P (ARM Target): cleaned up ARM deployment successfully: %s", component.Component.Name)
			ret[component.Component.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.Deleted,
				Message: "",
			}
		}
	}
	return ret, nil
}

func (r *ArmTargetProvider) cleanUpDeployment(ctx context.Context, deploymentName string, deployment *ArmDeployment) error {
	pollerResp, err := r.ResourceGroupClient.BeginDelete(ctx, deployment.ResourceGroup, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
func (r *ArmTargetProvider) analyzeDeployment(ctx context.Context, deploymentName string, deployment *ArmDeployment) error {
	template, err := deployment.Template.GetJson()
	if err != nil {
		return fmt.Errorf("cannot get template json: %v", err)
	}

}
func (r *ArmTargetProvider) createDeployment(ctx context.Context, deploymentName string, deployment *ArmDeployment, completeMode bool) (*armresources.DeploymentExtended, error) {
	template, err := deployment.Template.GetJson()
	if err != nil {
		return nil, fmt.Errorf("cannot get template json: %v", err)
	}
	params := map[string]interface{}{}
	if !deployment.Parameters.IsEmpty() {
		params, err = deployment.Parameters.GetJson()
		if err != nil {
			return nil, fmt.Errorf("cannot get parameters json: %v", err)
		}
	}
	deploymentMode := armresources.DeploymentModeIncremental
	if completeMode {
		deploymentMode = armresources.DeploymentModeComplete
	}
	deploymentPollerResp, err := r.DeploymentsClient.BeginCreateOrUpdate(
		ctx,
		deployment.ResourceGroup,
		deploymentName,
		armresources.Deployment{
			Properties: &armresources.DeploymentProperties{
				Template:   template,
				Parameters: params,
				Mode:       to.Ptr(deploymentMode),
			},
		},
		nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create deployment: %v", err)
	}

	resp, err := deploymentPollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get the create deployment future respone: %v", err)
	}

	return &resp.DeploymentExtended, nil
}

func (r *ArmTargetProvider) ensureResourceGroup(ctx context.Context, subscriptionId string, location string, resourceGroup string) (*armresources.ResourceGroup, error) {

	param := armresources.ResourceGroup{
		Location: to.Ptr(location),
	}

	resourceGroupResp, err := r.ResourceGroupClient.CreateOrUpdate(ctx, resourceGroup, param, nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func toARMTargetProviderConfig(config providers.IProviderConfig) (ArmTargetProviderConfig, error) {
	ret := ArmTargetProviderConfig{}
	if config == nil {
		return ret, errors.New("ARMTargetProviderConfig is null")
	}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (*ArmTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{},
			OptionalProperties:    []string{"yaml", "resource"},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
			ChangeDetectionProperties: []model.PropertyDesc{
				{Name: "yaml", IgnoreCase: false, SkipIfMissing: true},
				{Name: "resource", IgnoreCase: false, SkipIfMissing: true},
			},
		},
	}
}
