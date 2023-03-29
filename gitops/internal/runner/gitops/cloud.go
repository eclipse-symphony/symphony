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
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type cloudDeploymentRunner struct {
	gitops              *models.CloudGitOpsResource
	interval            time.Duration
	repoClient          clients.RepoClient
	ctx                 context.Context
	cancel              context.CancelFunc
	log                 logger.Logger
	dc                  *armresources.DeploymentsClient
	latestTemplateSha   string
	latestParametersSha string
	done                chan struct{}
	onceStart           sync.Once
	onceStop            sync.Once
}

func NewCloudDeploymentRunner(ctx context.Context, gitops *models.CloudGitOpsResource, repoClient clients.RepoClient) (runner.Runner, error) {

	cred, err := azidentity.NewDefaultAzureCredential(&defaultAzCredentialOptions)
	if err != nil {
		return nil, err
	}

	deploymentClient, err := armresources.NewDeploymentsClient(gitops.GetSubscription(), cred, &armclientOptions)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	return &cloudDeploymentRunner{
		ctx:        ctx,
		cancel:     cancel,
		gitops:     gitops,
		repoClient: repoClient,
		interval:   gitops.Properties.GetInterval(),
		dc:         deploymentClient,
		log:        logger.NewLogger(ctx, "gitops/cloudDeployment").WithField("gitops", gitops.Name),
		done:       make(chan struct{}),
	}, nil
}

func (g *cloudDeploymentRunner) GetId() string {
	return g.gitops.Id
}

func (g *cloudDeploymentRunner) Start() {
	g.onceStart.Do(g.start)
}

func (g *cloudDeploymentRunner) start() {
	g.log.Info("Starting GitOpsDeploymentRunner")
	go g.run()
}

func (g *cloudDeploymentRunner) Stop() {
	g.onceStop.Do(g.stop)
}

func (g *cloudDeploymentRunner) stop() {
	g.log.Infof("Stopping GitOpsDeploymentRunner")
	g.cancel()
}

func (g *cloudDeploymentRunner) Done() <-chan struct{} {
	return g.done
}

func (g *cloudDeploymentRunner) run() {
	defer close(g.done)

	ticker := time.NewTicker(g.interval)
	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			g.log.Info("Running GitOpsDeploymentRunner")
			err := g.fetchAndApply()
			if err != nil {
				g.log.WithError(err).Error("Error running GitOpsDeploymentRunner")
			}
		}
	}
}

func (g *cloudDeploymentRunner) fetchAndApply() error {
	//TODO: Get the template and parameters from the repo at the same time
	g.log.Info("Fetching template and parameters from repo")
	template, err := g.repoClient.GetContent(g.ctx, g.gitops.Properties.TemplatePath)
	if err != nil {
		return err
	}

	parameters, err := g.repoClient.GetContent(g.ctx, g.gitops.Properties.ParametersPath)
	if err != nil {
		return err
	}

	// Check if the template and parameters have changed
	if *template.SHA == g.latestTemplateSha && *parameters.SHA == g.latestParametersSha {
		g.log.Info("Template and parameters have not changed, skipping deployment")
		return nil
	}
	templateContent, err := template.GetContent()
	if err != nil {
		return err
	}

	parametersContent, err := parameters.GetContent()
	if err != nil {
		return err
	}

	err = g.deployToAzure(templateContent, parametersContent)
	if err != nil {
		return err
	}

	// Update the latest template and parameters
	g.latestTemplateSha = *template.SHA
	g.latestParametersSha = *parameters.SHA

	return nil
}

func (g *cloudDeploymentRunner) deployToAzure(deploymentTemplate string, parameters string) error {
	// Define the deployment properties
	hash := sha1.New()
	hash.Write([]byte(g.gitops.Id))
	deploymentName := fmt.Sprintf("gitops-deployment-%x", hash.Sum(nil))
	resourceGroupName := g.gitops.GetResourceGroup()
	parametersMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(parameters), &parametersMap)
	if err != nil {
		return err
	}
	templateMap := map[string]interface{}{}
	err = json.Unmarshal([]byte(deploymentTemplate), &templateMap)
	if err != nil {
		return err
	}
	// Create the deployment
	deployment := armresources.Deployment{
		Properties: &armresources.DeploymentProperties{
			Template:   templateMap,
			Parameters: parametersMap,
			Mode:       to.Ptr(g.gitops.Properties.Mode),
		},
	}

	ctx, cancel := context.WithTimeout(g.ctx, 30*time.Minute)
	defer cancel()
	g.log.Infof("Creating deployment %s in resource group %s", deploymentName, resourceGroupName)
	_, err = g.dc.BeginCreateOrUpdate(ctx, resourceGroupName, deploymentName, deployment, nil)
	if err != nil {
		return err
	}

	g.log.Info("Deployment created successfully!")
	return nil
}
