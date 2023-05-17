package gitops

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"dev.azure.com/msazure/One/_git/symphony/gitops/internal/runner"
	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/clients"
	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/logger"
	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/models"
	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/serving"
	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/utils"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/extendedlocation/armextendedlocation"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
)

type edgeRunner struct {
	gitops     *models.EdgeGitOpsResource
	interval   time.Duration
	repoClient clients.RepoClient
	ctx        context.Context
	cancel     context.CancelFunc
	log        logger.Logger
	dc         *armresources.DeploymentsClient
	clc        *armextendedlocation.CustomLocationsClient
	shas       map[string]string
	done       chan struct{}
	onceStart  sync.Once
	onceStop   sync.Once
}

func NewEdgeRunner(ctx context.Context, gitops *models.EdgeGitOpsResource, repoClient clients.RepoClient) (runner.Runner, error) {
	cred, err := azidentity.NewDefaultAzureCredential(&defaultAzCredentialOptions)
	if err != nil {
		return nil, err
	}

	subscriptionId := utils.GetSubscriptionFromResourceId(gitops.Properties.ExtendedLocationId)

	dc, err := armresources.NewDeploymentsClient(subscriptionId, cred, &armclientOptions)
	if err != nil {
		return nil, err
	}

	clc, err := armextendedlocation.NewCustomLocationsClient(subscriptionId, cred, &armclientOptions)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	return &edgeRunner{
		ctx:        ctx,
		cancel:     cancel,
		gitops:     gitops,
		repoClient: repoClient,
		interval:   gitops.Properties.GetInterval(),
		dc:         dc,
		clc:        clc,
		log:        logger.NewLogger(ctx, "gitops/edgeDeployment").WithField("gitops", gitops.Name),
		done:       make(chan struct{}),
		shas:       make(map[string]string, 0),
	}, nil
}

func (e *edgeRunner) GetId() string {
	return e.gitops.Id
}

func (e *edgeRunner) Start() {
	e.onceStart.Do(e.start)
}

func (e *edgeRunner) Stop() {
	e.onceStop.Do(e.stop)
}

func (e *edgeRunner) Done() <-chan struct{} {
	return e.done
}

func (e *edgeRunner) start() {
	e.log.Info("Starting edge deployment runner")
	go e.run()
}

func (e *edgeRunner) run() {
	defer close(e.done)
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()
	for {
		select {
		case <-e.ctx.Done():
			e.log.Info("Stopping edge deployment runner")
			return
		case <-ticker.C:
			e.log.Info("Running edge deployment")
			err := e.fetchAndApply()
			if err != nil {
				e.log.WithError(err).Error("Failed to run edge deployment")
			}
		}
	}
}

func (e *edgeRunner) stop() {
	e.cancel()
}

func (e *edgeRunner) fetchAndApply() error {
	e.log.Info("Fetching Edge Template and Parameters from repo")
	needsUpdate := false
	gitopsTemplates := make(map[string]models.GitOpsEdgeTemplate, 0)
	gitopsParameters := make(map[string]models.GitOpsEdgeParameters, 0)
	hashes := make(map[string]string, 0)
	for _, stage := range e.gitops.Properties.DeploymentScheme.Stages {
		solution, err := e.repoClient.GetContent(e.ctx, stage.Template.Path)
		if err != nil {
			return err
		}
		parameters, err := e.repoClient.GetContent(e.ctx, stage.Parameters.Path)
		if err != nil {
			return err
		}
		needsUpdate = needsUpdate || e.getSha(stage.Template.Path) != solution.GetSHA() || e.getSha(stage.Parameters.Path) != parameters.GetSHA()
		hashes[stage.Template.Path] = solution.GetSHA()
		hashes[stage.Parameters.Path] = parameters.GetSHA()
		solutionContent, err := solution.GetContent()
		if err != nil {
			return err
		}
		template := models.GitOpsEdgeTemplate{
			Template: make(map[string]interface{}),
			Name:     stage.Template.Name,
		}
		err = json.Unmarshal([]byte(solutionContent), &template.Template)
		if err != nil {
			return err
		}
		gitopsTemplates[stage.Name] = template

		parameterContent, err := parameters.GetContent()
		if err != nil {
			return err
		}
		parameter := models.GitOpsEdgeParameters{
			Parameters: make(map[string]interface{}),
			Name:       stage.Parameters.Name,
		}
		err = json.Unmarshal([]byte(parameterContent), &parameter.Parameters)
		if err != nil {
			return err
		}
		gitopsParameters[stage.Name] = parameter
	}
	if !needsUpdate {
		e.log.Info("No changes detected, skipping deployment")
		return nil
	}
	err := e.apply(gitopsTemplates, gitopsParameters)
	if err != nil {
		return err
	}
	for path, sha := range hashes {
		e.shas[path] = sha
	}
	return nil
}

func (e *edgeRunner) getSha(path string) string {
	return e.shas[path]
}
func (e *edgeRunner) apply(gitOpsEdgeTemplates map[string]models.GitOpsEdgeTemplate, gitOpsEdgeParameters map[string]models.GitOpsEdgeParameters) error {
	hash := sha1.New()
	hash.Write([]byte(e.gitops.Id))
	deploymentName := fmt.Sprintf("gitops-edge-%x", hash.Sum(nil))

	// get the resource group from the extended locations hostResourceId
	e.log.Info("Getting extended location host resource group")
	extendedLocation, err := e.clc.Get(e.ctx, utils.GetResourceGroupFromResourceId(
		e.gitops.Properties.ExtendedLocationId),
		utils.GetResourceNameFromResourceId(e.gitops.Properties.ExtendedLocationId), nil)
	if err != nil {
		return err
	}
	resourceGroupName := utils.GetResourceGroupFromResourceId(*extendedLocation.Properties.HostResourceID)

	resources := make([]map[string]interface{}, 0)

	for _, gitOpsEdgeTemplate := range gitOpsEdgeTemplates {
		resources = append(resources, buildSolution(gitOpsEdgeTemplate, extendedLocation.CustomLocation))
	}
	// we would also append the parameters but symphony doesn't support it yet

	resources = append(resources, buildInstance(e.gitops, gitOpsEdgeTemplates, gitOpsEdgeParameters, extendedLocation.CustomLocation))

	deployment := armresources.Deployment{
		Properties: &armresources.DeploymentProperties{
			Template: map[string]interface{}{
				"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
				"contentVersion": "1.0.0.0",
				"parameters":     map[string]interface{}{},
				"variables":      map[string]interface{}{},
				"resources":      resources,
			},
			Mode: to.Ptr(e.gitops.Properties.Mode),
		},
	}

	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Minute)
	defer cancel()
	e.log.Infof("Creating edge deployment %s in resource group %s", deploymentName, resourceGroupName)
	_, err = e.dc.BeginCreateOrUpdate(ctx, resourceGroupName, deploymentName, deployment, nil)
	if err != nil {
		return err
	}

	e.log.Info("Deployment created successfully!")
	return nil
}

func buildSolution(gitOpsEdgeTemplate models.GitOpsEdgeTemplate, exl armextendedlocation.CustomLocation) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": serving.ApiVersion,
		"location":   exl.Location,
		"extendedLocation": map[string]interface{}{
			"name": exl.ID,
			"type": "CustomLocation",
		},
		"type":       fmt.Sprintf("%s/Solutions", serving.ProviderNamespace),
		"name":       gitOpsEdgeTemplate.Name,
		"properties": gitOpsEdgeTemplate.Template,
	}
}

func buildInstance(gitops *models.EdgeGitOpsResource, gitOpsEdgeTemplates map[string]models.GitOpsEdgeTemplate, gitOpsEdgeParameters map[string]models.GitOpsEdgeParameters, exl armextendedlocation.CustomLocation) map[string]interface{} {
	var stage models.GitOpsStage
	if len(gitops.Properties.DeploymentScheme.Stages) != 0 {
		stage = gitops.Properties.DeploymentScheme.Stages[0]
	}

	instance := model.InstanceSpec{
		Scope:    gitops.Properties.DeploymentScheme.Scope,
		Name:     stage.Name,
		Solution: gitOpsEdgeTemplates[stage.Name].Name,
		Target:   stage.TargetRef,
	}

	dependencies := make([]string, 0)
	for _, tpl := range gitOpsEdgeTemplates {
		dependencies = append(dependencies, tpl.Name)
	}

	instanceMap := map[string]interface{}{}
	instanceString, _ := json.Marshal(instance)
	json.Unmarshal(instanceString, &instanceMap)
	//HACK: delete instance.name because it's not supported yet
	delete(instanceMap, "name")
	return map[string]interface{}{
		"apiVersion": serving.ApiVersion,
		"location":   exl.Location,
		"extendedLocation": map[string]interface{}{
			"name": exl.ID,
			"type": "CustomLocation",
		},
		"type":       fmt.Sprintf("%s/Instances", serving.ProviderNamespace),
		"name":       gitops.Name,
		"properties": instanceMap,
		"dependsOn":  dependencies,
	}
}
