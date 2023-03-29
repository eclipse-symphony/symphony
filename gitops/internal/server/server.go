package server

import (
	"context"
	"encoding/json"
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
		Server:          &fasthttp.Server{},
		r:               router.New(),
		manager:         manager.NewManager(),
		gitopsRegistrar: clients.DefaultGitOpsRegistrar,
		log:             logger.NewLogger(context.Background(), "Server"),
	}
}

func (s *Server) registerRoutes() {
	s.initializer.Do(func() {
		// index rout
		s.r.GET("/", s.index)
		s.r.GET(serving.HealthzEndpoint, s.healthz)
		s.r.GET(serving.ReadyzEndpoint, s.readyz)
		s.r.PUT(serving.RepoURLEndpoint, s.upsertRepo)
		s.r.GET(serving.RepoURLEndpoint, s.getRepo)
		s.r.PUT(serving.CloudGitOpsEndpoint, s.upsertCloudGitOps)
		s.r.DELETE(serving.CloudGitOpsEndpoint, s.stopCloudGitOps)
		s.r.PUT(serving.EdgeGitOpsEndpoint, s.upsertEdgeGitOps)
		s.r.DELETE(serving.EdgeGitOpsEndpoint, s.stopEdgeGitOps)
		s.r.POST(serving.CommitEndpoint, s.handleCommit)
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
	data := make(map[string]interface{})
	data["message"] = "GitOps service is running."
	s.jsonResponse(ctx, fasthttp.StatusOK, data)
}

func (s *Server) readyz(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (s *Server) upsertRepo(ctx *fasthttp.RequestCtx) {
	repoResource := &models.RepoRequest{}
	if err := json.Unmarshal(ctx.PostBody(), repoResource); err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusBadRequest, err)
		return
	}

	// Register repo
	if err := s.gitopsRegistrar.RegisterRepo(repoResource); err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}

	s.jsonResponse(ctx, fasthttp.StatusOK, repoResource)
}

// TODO: Possibly remove. currently for debugging
func (s *Server) getRepo(ctx *fasthttp.RequestCtx) {
	subscriptionId := ctx.UserValue("subscriptionId").(string)
	resourceGroup := ctx.UserValue("resourceGroup").(string)
	repoName := ctx.UserValue("repoName").(string)
	client, err := s.gitopsRegistrar.GetRepoClient(subscriptionId, resourceGroup, repoName)
	if err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}

	repo := client.GetResource()
	s.jsonResponse(ctx, fasthttp.StatusOK, &repo)

}

func (s *Server) upsertCloudGitOps(ctx *fasthttp.RequestCtx) {
	deployUtilizationResource := &models.DeploymentUtilationRequest{}
	if err := json.Unmarshal(ctx.PostBody(), deployUtilizationResource); err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusBadRequest, err)
		return
	}

	// get repo client
	repoClient, err := s.gitopsRegistrar.GetRepoClient(deployUtilizationResource.GetSubscription(), deployUtilizationResource.GetResourceGroup(), deployUtilizationResource.GetAzRepoShortName())
	if err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}

	if err := s.gitopsRegistrar.RegisterDeploymentUtilization(deployUtilizationResource); err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}

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
	s.manager.AddRunner(runner)

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
	repoClient, err := s.gitopsRegistrar.GetRepoClient(edgeUtilizationResource.GetSubscription(), edgeUtilizationResource.GetResourceGroup(), edgeUtilizationResource.GetAzRepoShortName())
	if err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}

	if err := s.gitopsRegistrar.RegisterEdgeUtilization(edgeUtilizationResource); err != nil {
		s.log.Error(err)
		s.jsonErrorResponse(ctx, fasthttp.StatusInternalServerError, err)
		return
	}

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
	s.manager.AddRunner(runner)

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
	repoClient, err := registrar.GetRepoClient(subscriptionId, resourceGroup, azRepoName)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	err = repoClient.CommitFiles(ctx, req.Files, req.CommitMessage)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}

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
