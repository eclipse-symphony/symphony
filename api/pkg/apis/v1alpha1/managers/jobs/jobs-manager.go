/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

const Scheduled = "Scheduled"

type JobsManager struct {
	managers.Manager
	PersistentStateProvider states.IStateProvider
	VolatileStateProvider   states.IStateProvider
	apiClient               utils.ApiClient
	interval                int32
	user                    string
	password                string
}

type LastSuccessTime struct {
	Time time.Time `json:"time"`
}

func (s *JobsManager) Init(vContext *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(vContext, config, providers)
	if err != nil {
		return err
	}

	volatilestateprovider, err := managers.GetVolatileStateProvider(config, providers)
	if err == nil {
		s.VolatileStateProvider = volatilestateprovider
	} else {
		return err
	}
	persistentStateProvider, err := managers.GetPersistentStateProvider(config, providers)
	if err == nil {
		s.PersistentStateProvider = persistentStateProvider
	} else {
		return err
	}
	if utils.ShouldUseUserCreds() {
		user, err := utils.GetString(s.Manager.Config.Properties, "user")
		if err != nil {
			return err
		}
		s.user = user
		if s.user == "" {
			return v1alpha2.NewCOAError(nil, "user is required", v1alpha2.BadConfig)
		}
		password, err := utils.GetString(s.Manager.Config.Properties, "password")
		if err != nil {
			return err
		}
		s.password = password
	}

	s.interval = utils.ReadInt32(s.Manager.Config.Properties, "interval", 0)

	s.apiClient, err = utils.GetApiClient()
	if err != nil {
		return err
	}
	return nil
}

func (s *JobsManager) Enabled() bool {
	return s.Config.Properties["poll.enabled"] == "true" || s.Config.Properties["schedule.enabled"] == "true"
}

func (s *JobsManager) pollObjects() []error {
	context, span := observability.StartSpan("Job Manager", context.Background(), &map[string]string{
		"method": "pollObjects",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	if s.interval == 0 {
		return nil
	}

	var instances []model.InstanceState
	instances, err = s.apiClient.GetInstancesForAllNamespaces(context, s.user, s.password)
	if err != nil {
		log.ErrorfCtx(context, " M (Job): error getting instances: %s", err.Error())
		return []error{err}
	}
	for _, instance := range instances {
		var entry states.StateEntry
		entry, err = s.VolatileStateProvider.Get(context, states.GetRequest{
			ID: "i_" + instance.ObjectMeta.Name,
			Metadata: map[string]interface{}{
				"namespace": instance.ObjectMeta.Namespace,
			},
		})
		needsPub := true
		if err == nil {
			var stamp LastSuccessTime
			if stamp, err = getLastSuccessTime(entry.Body); err == nil {
				if time.Since(stamp.Time) > time.Duration(s.interval)*time.Second { //TODO: compare object hash as well?
					needsPub = true
				} else {
					needsPub = false
				}
			}
		}
		if needsPub {
			s.Context.Publish("job", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": "instance",
				},
				Body: v1alpha2.JobData{
					Id:     instance.ObjectMeta.Name,
					Action: v1alpha2.JobUpdate,
					Scope:  instance.ObjectMeta.Namespace,
				},
				Context: context,
			})
		}
	}
	var targets []model.TargetState
	targets, err = s.apiClient.GetTargetsForAllNamespaces(context, s.user, s.password)
	if err != nil {
		log.ErrorfCtx(context, " M (Job): error getting targets: %s", err.Error())
		return []error{err}
	}
	for _, target := range targets {
		var entry states.StateEntry
		entry, err = s.VolatileStateProvider.Get(context, states.GetRequest{
			ID: "t_" + target.ObjectMeta.Name,
			Metadata: map[string]interface{}{
				"namespace": target.ObjectMeta.Namespace,
			},
		})
		needsPub := true
		if err == nil {
			var stamp LastSuccessTime
			if stamp, err = getLastSuccessTime(entry.Body); err == nil {
				if time.Since(stamp.Time) > time.Duration(s.interval)*time.Second { //TODO: compare object hash as well?
					needsPub = true
				} else {
					needsPub = false
				}
			}
		}
		if needsPub {
			s.Context.Publish("job", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": "target",
				},
				Body: v1alpha2.JobData{
					Id:     target.ObjectMeta.Name,
					Action: v1alpha2.JobUpdate,
					Scope:  target.ObjectMeta.Namespace,
				},
				Context: context,
			})
		}
	}

	return nil
}
func (s *JobsManager) Poll() []error {
	// TODO: do these in parallel?
	if s.Config.Properties["poll.enabled"] == "true" {
		errors := s.pollObjects()
		if len(errors) > 0 {
			return errors
		}
	}
	if s.Config.Properties["schedule.enabled"] == "true" {
		errors := s.pollSchedules()
		if len(errors) > 0 {
			return errors
		}
	}
	return nil
}

func (s *JobsManager) pollSchedules() []error {
	context, span := observability.StartSpan("Job Manager", context.Background(), &map[string]string{
		"method": "pollSchedules",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	//TODO: use filters and continue tokens
	var list []states.StateEntry
	list, _, err = s.PersistentStateProvider.List(context, states.ListRequest{
		Metadata: map[string]interface{}{
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": Scheduled,
		},
	})
	if err != nil {
		return []error{err}
	}

	for _, entry := range list {
		var activationData v1alpha2.ActivationData
		entryData, _ := json.Marshal(entry.Body)
		err = json.Unmarshal(entryData, &activationData)
		if err != nil {
			return []error{err}
		}
		if activationData.Schedule != "" {
			var fire bool
			fire, err = activationData.ShouldFireNow()
			if err != nil {
				return []error{err}
			}
			if fire {
				activationData.Schedule = ""
				err = s.PersistentStateProvider.Delete(context, states.DeleteRequest{
					ID: entry.ID,
					Metadata: map[string]interface{}{
						"namespace": activationData.Namespace,
						"group":     model.WorkflowGroup,
						"version":   "v1",
						"resource":  Scheduled,
					},
				})
				if err != nil {
					return []error{err}
				}
				s.Context.Publish("trigger", v1alpha2.Event{
					Body:    activationData,
					Context: context,
				})
			}
		}
	}
	return nil
}

func (s *JobsManager) Reconcil() []error {
	return nil
}
func (s *JobsManager) HandleHeartBeatEvent(ctx context.Context, event v1alpha2.Event) error {
	ctx, span := observability.StartSpan("Job Manager", ctx, &map[string]string{
		"method": "HandleHeartBeatEvent",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	var heartbeat v1alpha2.HeartBeatData
	jData, _ := json.Marshal(event.Body)
	err = json.Unmarshal(jData, &heartbeat)
	if err != nil {
		err = v1alpha2.NewCOAError(nil, "event body is not a heart beat", v1alpha2.BadRequest)
		return err
	}

	namespace, ok := event.Metadata["namespace"]
	if !ok {
		namespace = "default"
	}
	// TODO: the heart beat data should contain a "finished" field so data can be cleared
	log.DebugfCtx(ctx, " M (Job): handling heartbeat h_%s", heartbeat.JobId)
	_, err = s.VolatileStateProvider.Upsert(ctx, states.UpsertRequest{
		Value: states.StateEntry{
			ID:   "h_" + heartbeat.JobId,
			Body: heartbeat,
		},
		Metadata: map[string]interface{}{
			"namespace": namespace,
		},
	})
	return err
}

func (s *JobsManager) DelayOrSkipJob(ctx context.Context, namespace string, objectType string, job v1alpha2.JobData) error {
	ctx, span := observability.StartSpan("Job Manager", ctx, &map[string]string{
		"method": "DelayOrSkipJob",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	key := "h_" + job.Id
	if objectType == "target" {
		key = fmt.Sprintf("h_%s-%s", "target-runtime", job.Id)
	}
	//check if a manager is working on the job
	var entry states.StateEntry
	entry, err = s.VolatileStateProvider.Get(ctx, states.GetRequest{
		ID: key,
		Metadata: map[string]interface{}{
			"namespace": namespace,
		},
	})
	if err != nil {
		if !v1alpha2.IsNotFound(err) {
			log.ErrorfCtx(ctx, " M (Job): error getting heartbeat %s: %s", key, err.Error())
			return err
		}
		log.DebugfCtx(ctx, " M (Job): found heartbeat %s, entry: %+v", key, entry)
		return nil // no heartbeat
	}
	var heartbeat v1alpha2.HeartBeatData
	jData, _ := json.Marshal(entry.Body)
	err = json.Unmarshal(jData, &heartbeat)
	if err != nil {
		return err
	}
	if time.Since(heartbeat.Time) > time.Duration(60)*time.Second { //TODO: make this configurable
		// heartbeat is too old
		return nil
	}

	if job.Action == v1alpha2.JobDelete && heartbeat.Action == v1alpha2.HeartBeatUpdate {
		err = v1alpha2.NewCOAError(nil, "delete job is delayed", v1alpha2.Delayed)
		return err
	}
	err = v1alpha2.NewCOAError(nil, "existing job in progress", v1alpha2.Untouched)
	return err
}
func (s *JobsManager) HandleScheduleEvent(ctx context.Context, event v1alpha2.Event) error {
	ctx, span := observability.StartSpan("Job Manager", ctx, &map[string]string{
		"method": "HandleScheduleEvent",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	var activationData v1alpha2.ActivationData
	jData, _ := json.Marshal(event.Body)
	err = json.Unmarshal(jData, &activationData)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "event body is not a activation data", v1alpha2.BadRequest)
	}
	key := fmt.Sprintf("sch_%s-%s", activationData.Campaign, activationData.Activation)
	_, err = s.PersistentStateProvider.Upsert(ctx, states.UpsertRequest{
		Value: states.StateEntry{
			ID:   key,
			Body: activationData,
		},
		Metadata: map[string]interface{}{
			"namespace": activationData.Namespace,
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  Scheduled,
		},
	})
	return err
}
func (s *JobsManager) HandleJobEvent(ctx context.Context, event v1alpha2.Event) error {
	ctx, span := observability.StartSpan("Job Manager", ctx, &map[string]string{
		"method": "HandleJobEvent",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	namespace := model.ReadProperty(event.Metadata, "namespace", nil)
	if namespace == "" {
		namespace = "default"
	}

	if objectType, ok := event.Metadata["objectType"]; ok {
		var job v1alpha2.JobData
		jData, _ := json.Marshal(event.Body)
		err = json.Unmarshal(jData, &job)
		if err != nil {
			return v1alpha2.NewCOAError(nil, "event body is not a job", v1alpha2.BadRequest)
		}

		err = s.DelayOrSkipJob(ctx, namespace, objectType, job)
		if err != nil {
			return err
		}

		switch objectType {
		case "instance":
			log.DebugfCtx(ctx, " M (Job): handling instance job %s", job.Id)
			instanceName := job.Id
			var instance model.InstanceState
			//get intance
			instance, err = s.apiClient.GetInstance(ctx, instanceName, namespace, s.user, s.password)
			if err != nil {
				log.ErrorfCtx(ctx, " M (Job): error getting instance %s, namespace: %s: %s", instanceName, namespace, err.Error())
				return err //TODO: instance is gone
			}

			//get solution
			var solution model.SolutionState
			solutionName := api_utils.ReplaceSeperator(instance.Spec.Solution)
			solution, err = s.apiClient.GetSolution(ctx, solutionName, namespace, s.user, s.password)
			if err != nil {
				solution = model.SolutionState{
					ObjectMeta: model.ObjectMeta{
						Name:      instance.Spec.Solution,
						Namespace: namespace,
					},
					Spec: &model.SolutionSpec{
						Components: make([]model.ComponentSpec, 0),
					},
				}
			}

			//get targets
			var targets []model.TargetState
			targets, err = s.apiClient.GetTargets(ctx, namespace, s.user, s.password)
			if err != nil {
				targets = make([]model.TargetState, 0)
			}

			//get target candidates
			targetCandidates := utils.MatchTargets(instance, targets)

			//create deployment spec
			var deployment model.DeploymentSpec
			deployment, err = utils.CreateSymphonyDeployment(instance, solution, targetCandidates, nil, namespace)
			if err != nil {
				log.ErrorfCtx(ctx, " M (Job): error creating deployment spec for instance %s: %s", instanceName, err.Error())
				return err
			}

			//call api
			switch job.Action {
			case v1alpha2.JobUpdate:
				_, err = s.apiClient.Reconcile(ctx, deployment, false, namespace, s.user, s.password)
				if err != nil {
					log.ErrorfCtx(ctx, " M (Job): error reconciling instance %s: %s", instanceName, err.Error())
					return err
				} else {
					s.VolatileStateProvider.Upsert(ctx, states.UpsertRequest{
						Value: states.StateEntry{
							ID: "i_" + instance.ObjectMeta.Name,
							Body: LastSuccessTime{
								Time: time.Now().UTC(),
							},
						},
						Metadata: map[string]interface{}{
							"namespace": namespace,
						},
					})
				}
			case v1alpha2.JobDelete:
				_, err = s.apiClient.Reconcile(ctx, deployment, true, namespace, s.user, s.password)
				if err != nil {
					return err
				} else {
					return s.apiClient.DeleteInstance(ctx, deployment.Instance.ObjectMeta.Name, namespace, s.user, s.password)
				}
			default:
				return v1alpha2.NewCOAError(nil, "unsupported action", v1alpha2.BadRequest)
			}
		case "target":
			var target model.TargetState
			targetName := job.Id
			target, err = s.apiClient.GetTarget(ctx, targetName, namespace, s.user, s.password)
			if err != nil {
				return err
			}
			var deployment model.DeploymentSpec
			deployment, err = utils.CreateSymphonyDeploymentFromTarget(target, namespace)
			if err != nil {
				return err
			}
			switch job.Action {
			case v1alpha2.JobUpdate:
				_, err = s.apiClient.Reconcile(ctx, deployment, false, namespace, s.user, s.password)
				if err != nil {
					return err
				} else {
					// TODO: how to handle status updates?
					s.VolatileStateProvider.Upsert(ctx, states.UpsertRequest{
						Value: states.StateEntry{
							ID: "t_" + targetName,
							Body: LastSuccessTime{
								Time: time.Now().UTC(),
							},
						},
						Metadata: map[string]interface{}{
							"namespace": namespace,
						},
					})
				}
			case v1alpha2.JobDelete:
				_, err = s.apiClient.Reconcile(ctx, deployment, true, namespace, s.user, s.password)
				if err != nil {
					return err
				} else {
					return s.apiClient.DeleteTarget(ctx, targetName, namespace, s.user, s.password)
				}
			default:
				return v1alpha2.NewCOAError(nil, "unsupported action", v1alpha2.BadRequest)
			}
		case "deployment":
			log.InfofCtx(ctx, " M (Job): handling deployment job %s, action: %s", job.Id, job.Action)
			log.InfofCtx(ctx, " M (Job): deployment spec: %s", string(job.Data))

			var deployment *model.DeploymentSpec
			deployment, err = model.ToDeployment(job.Data)
			if err != nil {
				return err
			}
			if job.Action == v1alpha2.JobUpdate {
				_, err = s.apiClient.Reconcile(ctx, *deployment, false, namespace, s.user, s.password)
				if err != nil {
					return err
				} else {
					// TODO: how to handle status updates?
					s.VolatileStateProvider.Upsert(ctx, states.UpsertRequest{
						Value: states.StateEntry{
							ID: "d_" + deployment.Instance.ObjectMeta.Name,
							Body: LastSuccessTime{
								Time: time.Now().UTC(),
							},
						},
						Metadata: map[string]interface{}{
							"namespace": namespace,
						},
					})
				}
			}
			if job.Action == v1alpha2.JobDelete {
				_, err = s.apiClient.Reconcile(ctx, *deployment, true, namespace, s.user, s.password)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getLastSuccessTime(body interface{}) (LastSuccessTime, error) {
	var lastSuccessTime LastSuccessTime
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &lastSuccessTime)
	if err != nil {
		return LastSuccessTime{}, err
	}
	return lastSuccessTime, nil
}
