package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"dev.azure.com/msazure/One/_git/symphony/gitops/internal/manager"
	"dev.azure.com/msazure/One/_git/symphony/gitops/internal/runner/gitops"
	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/clients"
	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/logger"
	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/models"
	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/serving"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type Server struct {
	*fasthttp.Server
	r               *router.Router
	initializer     sync.Once
	manager         manager.Manager
	gitopsRegistrar clients.GitOpsRegistrar
	log             logger.Logger
}

func NewServer() *Server {
	return &Server{
		// Buffer size was arbitrarily chosen as the default 4k size wasn't large enough to receive the header from metarp
		// Some investigation should be done to determine the correct size
		Server: &fasthttp.Server{
			ReadBufferSize: 8192,
		},
		r:               router.New(),
		manager:         manager.NewManager(),
		gitopsRegistrar: clients.DefaultGitOpsRegistrar,
		log:             logger.NewLogger(context.Background(), "Server"),
	}
}

func (s *Server) registerRoutes() {
	s.initializer.Do(func() {
		s.r.GET("/", s.index)
		s.r.GET(serving.HealthzEndpoint, s.healthz)
		s.r.GET(serving.ReadyzEndpoint, s.readyz)

		s.r.GET(serving.RepoURLEndpoint, s.getRepo)

		s.r.POST(serving.RepoCreationValidateEndpoint, s.createValidateResource)
		s.r.POST(serving.ArmDeploymentCreationValidateEndpoint, s.createValidateResource)
		s.r.POST(serving.EdgeDeploymentCreationValidateEndpoint, s.createValidateResource)

		s.r.PUT(serving.RepoCreationBegin, s.upsertRepo)
		s.r.PUT(serving.ArmDeploymentCreationBegin, s.upsertArmGitOps)
		s.r.PUT(serving.EdgeDeploymentCreationBegin, s.upsertEdgeGitOps)

		s.r.PATCH(serving.RepoPatchBegin, s.upsertRepo)
		s.r.PATCH(serving.ArmDeploymentPatchBegin, s.upsertArmGitOps)
		s.r.PATCH(serving.EdgeDeploymentPatchBegin, s.upsertEdgeGitOps)

		s.r.DELETE(serving.ArmDeploymentDeletionBegin, s.stopCloudGitOps)
		s.r.DELETE(serving.EdgeDeploymentDeletionBegin, s.stopEdgeGitOps)

		s.r.POST(serving.CommitEndpoint, s.handleCommit)

		s.ErrorHandler = s.errorHandler
		s.Handler = s.r.Handler
	})
}

func (s *Server) Start() {
	s.log.Info("Starting server")
	s.registerRoutes()
	if err := s.ListenAndServe(":8080"); err != nil {
		s.log.Fatal(err)
	}
}

func (s *Server) Stop() {
	s.log.Info("Stopping server")
	if err := s.Shutdown(); err != nil {
		s.log.Fatal(err)
	}
}

func (s *Server) healthz(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (s *Server) index(ctx *fasthttp.RequestCtx) {
	version := os.Getenv("IMAGE_TAG")
	data := make(map[string]interface{})
	data["message"] = fmt.Sprintf("GitOps service is running. Version: %s", version)
	s.jsonResponse(ctx, fasthttp.StatusOK, data)
}

func (s *Server) readyz(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (s *Server) upsertRepo(ctx *fasthttp.RequestCtx) {
	repoResource := &models.RepoRequest{}
	s.log.Infof("Creating new repository from body: %s", string(ctx.PostBody()))
	if err := json.Unmarshal(ctx.PostBody(), repoResource); err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusBadRequest, err)
		return
	}
	// Register repo
	s.log.Info("Registering repo")
	if err := s.gitopsRegistrar.RegisterRepo(repoResource); err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}
	s.log.Infof("Registered repo successfully: %s", repoResource)
	s.jsonResponse(ctx, fasthttp.StatusOK, repoResource)
}

// TODO: Possibly remove. currently for debugging
func (s *Server) getRepo(ctx *fasthttp.RequestCtx) {
	subscriptionId := ctx.UserValue("subscriptionId").(string)
	resourceGroup := ctx.UserValue("resourceGroup").(string)
	repoName := ctx.UserValue("repoName").(string)
	s.log.Info("Getting repo client")
	repoClient, err := s.gitopsRegistrar.GetRepoClient(subscriptionId, resourceGroup, repoName)
	if err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}
	s.log.Infof("Successfully got repo: %s", repoClient)
	repo := repoClient.GetResource()
	s.jsonResponse(ctx, fasthttp.StatusOK, &repo)

}

// Boiler plate validation for resource creation
// TODO: Fill out with custom logic
func (s *Server) createValidateResource(ctx *fasthttp.RequestCtx) {
	s.log.Info("Validating Resource")
	s.log.Info("Successfully validated resource")
	s.jsonResponse(ctx, fasthttp.StatusOK, models.ValidateResponse{
		Status: "success"})
}

func (s *Server) upsertArmGitOps(ctx *fasthttp.RequestCtx) {
	s.log.Infof("Creating new arm deployment gitops from body: %s", string(ctx.PostBody()))
	deployUtilizationResource := &models.DeploymentUtilationRequest{}
	if err := json.Unmarshal(ctx.PostBody(), deployUtilizationResource); err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusBadRequest, err)
		return
	}

	// get repo client
	s.log.Info("Getting repo client")
	repoClient, err := s.gitopsRegistrar.GetRepoClient(deployUtilizationResource.GetSubscription(), deployUtilizationResource.GetResourceGroup(), deployUtilizationResource.GetAzRepoShortName())
	if err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}
	s.log.Infof("Successfully got repo: %s", repoClient)
	s.log.Info("Registering deployment utilization")
	if err := s.gitopsRegistrar.RegisterDeploymentUtilization(deployUtilizationResource); err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}
	s.log.Info("Registered deployment utilization successfully")
	s.log.Info("Creating new cloud deployment runner")
	runner, err := gitops.NewCloudDeploymentRunner(
		s.manager.Ctx(),
		deployUtilizationResource,
		repoClient,
	)

	if err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}
	s.log.Info("Created new cloud deployment runner successfully")
	s.log.Info("Adding runner to manager")
	s.manager.AddRunner(runner)
	s.log.Info("Added runner to manager successfully")
	s.jsonResponse(ctx, fasthttp.StatusOK, deployUtilizationResource)
}

func (s *Server) upsertEdgeGitOps(ctx *fasthttp.RequestCtx) {
	edgeUtilizationResource := &models.EdgeUtilizationRequest{}
	if err := json.Unmarshal(ctx.PostBody(), edgeUtilizationResource); err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusBadRequest, err)
		return
	}
	// get repo client
	s.log.Info("Getting repo client")
	repoClient, err := s.gitopsRegistrar.GetRepoClient(edgeUtilizationResource.GetSubscription(), edgeUtilizationResource.GetResourceGroup(), edgeUtilizationResource.GetAzRepoShortName())
	if err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}
	s.log.Infof("Successfully got repo: %s", repoClient)
	s.log.Info("Registering edge utilization")
	if err := s.gitopsRegistrar.RegisterEdgeUtilization(edgeUtilizationResource); err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}
	s.log.Info("Registered edge utilization successfully")
	s.log.Info("Creating new edge utilization runner")
	runner, err := gitops.NewEdgeRunner(
		s.manager.Ctx(),
		edgeUtilizationResource,
		repoClient,
	)
	if err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}
	s.log.Info("Created new edge utilization runner successfully")
	s.log.Info("Adding runner to manager")
	s.manager.AddRunner(runner)
	s.log.Info("Added runner to manager successfully")
	s.jsonResponse(ctx, fasthttp.StatusOK, edgeUtilizationResource)
}

func (s *Server) stopEdgeGitOps(ctx *fasthttp.RequestCtx) {
	subscriptionId := ctx.UserValue("subscriptionId").(string)
	resourceGroup := ctx.UserValue("resourceGroup").(string)
	repoName := ctx.UserValue("repoName").(string)
	utilizationName := ctx.UserValue("edgeUtilization").(string)
	// get repo client
	if _, err := s.gitopsRegistrar.GetRepoClient(subscriptionId, resourceGroup, repoName); err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusNotFound, err)
		return
	}
	utilization, err := s.gitopsRegistrar.GetEdgeUtilization(subscriptionId, resourceGroup, repoName, utilizationName)
	if err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}
	s.log.Infof("Removing edge utilization runner %s:", utilization.Id)
	s.manager.RemoveRunner(utilization.Id)

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (s *Server) stopCloudGitOps(ctx *fasthttp.RequestCtx) {
	subscriptionId := ctx.UserValue("subscriptionId").(string)
	resourceGroup := ctx.UserValue("resourceGroup").(string)
	repoName := ctx.UserValue("repoName").(string)
	utilizationName := ctx.UserValue("deploymentUtilization").(string)
	// get repo client
	if _, err := s.gitopsRegistrar.GetRepoClient(subscriptionId, resourceGroup, repoName); err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusNotFound, err)
		return
	}
	utilization, err := s.gitopsRegistrar.GetDeploymentUtilization(subscriptionId, resourceGroup, repoName, utilizationName)
	if err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}
	s.log.Infof("Removing cloud utilization runner %s:", utilization.Id)
	s.manager.RemoveRunner(utilization.Id)

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (s *Server) handleCommit(ctx *fasthttp.RequestCtx) {
	registrar := clients.DefaultGitOpsRegistrar
	// Get the request body
	var req models.GitCommitRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}
	subscriptionId := ctx.UserValue("subscriptionId").(string)
	resourceGroup := ctx.UserValue("resourceGroup").(string)
	azRepoName := ctx.UserValue("repoName").(string)
	s.log.Info("Getting repo client")
	repoClient, err := registrar.GetRepoClient(subscriptionId, resourceGroup, azRepoName)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	s.log.Infof("Successfully got repo: %s", repoClient)
	s.log.Infof("Handling commit: %s", req.CommitMessage)
	err = repoClient.CommitFiles(ctx, req.Files, req.CommitMessage)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	s.log.Info("Successfullly handled commit")

	ctx.SetStatusCode(fasthttp.StatusOK)
}

// json response helper
func (s *Server) jsonResponse(ctx *fasthttp.RequestCtx, statusCode int, body interface{}) {
	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("application/json")
	if err := json.NewEncoder(ctx).Encode(body); err != nil {
		s.log.Error(err)
	}
}

// json error response helper
func (s *Server) jsonErrorResponse(ctx *fasthttp.RequestCtx, statusCode int, err error) {
	s.jsonResponse(ctx, statusCode, map[string]string{"error": err.Error()})
}

func (s *Server) errorHandler(ctx *fasthttp.RequestCtx, err error) {
	s.log.Infof("Received error while handling request: %s", err.Error())
	s.jsonErrorResponse(ctx, 500, err)
}
