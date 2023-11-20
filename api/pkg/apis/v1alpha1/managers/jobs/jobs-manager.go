/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/azure/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

type JobsManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

type LastSuccessTime struct {
	Time time.Time `json:"time"`
}

func (s *JobsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}

	stateprovider, err := managers.GetStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
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
	defer observ_utils.CloseSpanWithError(span, err)

	baseUrl, err := utils.GetString(s.Manager.Config.Properties, "baseUrl")
	if err != nil {
		return []error{err}
	}
	user, err := utils.GetString(s.Manager.Config.Properties, "user")
	if err != nil {
		return []error{err}
	}
	password, err := utils.GetString(s.Manager.Config.Properties, "password")
	if err != nil {
		return []error{err}
	}
	interval := utils.ReadInt32(s.Manager.Config.Properties, "interval", 0)
	if interval == 0 {
		return nil
	}
	instances, err := utils.GetInstances(context, baseUrl, user, password)
	if err != nil {
		fmt.Println(err.Error())
		return []error{err}
	}
	for _, instance := range instances {
		entry, err := s.StateProvider.Get(context, states.GetRequest{
			ID: "i_" + instance.Id,
		})
		needsPub := true
		if err == nil {
			if stamp, ok := entry.Body.(LastSuccessTime); ok {
				if time.Since(stamp.Time) > time.Duration(interval)*time.Second { //TODO: compare object hash as well?
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
					Id:     instance.Id,
					Action: "UPDATE",
				},
			})
		}
	}
	targets, err := utils.GetTargets(context, baseUrl, user, password)
	if err != nil {
		fmt.Println(err.Error())
		return []error{err}
	}
	for _, target := range targets {
		entry, err := s.StateProvider.Get(context, states.GetRequest{
			ID: "t_" + target.Id,
		})
		needsPub := true
		if err == nil {
			var stamp LastSuccessTime
			jData, _ := json.Marshal(entry.Body)
			err = json.Unmarshal(jData, &stamp)
			if err == nil {
				if time.Since(stamp.Time) > time.Duration(interval)*time.Second { //TODO: compare object hash as well?
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
					Id:     target.Id,
					Action: "UPDATE",
				},
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
	defer observ_utils.CloseSpanWithError(span, err)

	//TODO: use filters and continue tokens
	list, _, err := s.StateProvider.List(context, states.ListRequest{})
	if err != nil {
		return []error{err}
	}

	for _, entry := range list {
		var activationData v1alpha2.ActivationData
		entryData, _ := json.Marshal(entry.Body)
		err := json.Unmarshal(entryData, &activationData)
		if err != nil {
			return []error{err}
		}
		if activationData.Schedule != nil {
			fire, err := activationData.Schedule.ShouldFireNow()
			if err != nil {
				return []error{err}
			}
			if fire {
				activationData.Schedule = nil
				err = s.StateProvider.Delete(context, states.DeleteRequest{
					ID: entry.ID,
				})
				if err != nil {
					return []error{err}
				}
				s.Context.Publish("trigger", v1alpha2.Event{
					Body: activationData,
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
	defer observ_utils.CloseSpanWithError(span, err)

	var heartbeat v1alpha2.HeartBeatData
	jData, _ := json.Marshal(event.Body)
	err = json.Unmarshal(jData, &heartbeat)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "event body is not a heart beat", v1alpha2.BadRequest)
	}
	// TODO: the heart beat data should contain a "finished" field so data can be cleared
	_, err = s.StateProvider.Upsert(ctx, states.UpsertRequest{
		Value: states.StateEntry{
			ID:   "h_" + heartbeat.JobId,
			Body: heartbeat,
		},
	})
	return err
}

func (s *JobsManager) DelayOrSkipJob(ctx context.Context, objectType string, job v1alpha2.JobData) error {
	ctx, span := observability.StartSpan("Job Manager", ctx, &map[string]string{
		"method": "DelayOrSkipJob",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, err)

	key := "h_" + job.Id
	if objectType == "target" {
		key = fmt.Sprintf("h_%s-%s", "target-runtime", job.Id)
	}
	//check if a manager is working on the job
	entry, err := s.StateProvider.Get(ctx, states.GetRequest{
		ID: key,
	})
	if err != nil {
		if !v1alpha2.IsNotFound(err) {
			return err
		}
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
	if job.Action == "delete" && heartbeat.Action == "update" {
		return v1alpha2.NewCOAError(nil, "delete job is delayed", v1alpha2.Delayed)
	}
	return v1alpha2.NewCOAError(nil, "existing job in progress", v1alpha2.Untouched)
}
func (s *JobsManager) HandleScheduleEvent(ctx context.Context, event v1alpha2.Event) error {
	ctx, span := observability.StartSpan("Job Manager", ctx, &map[string]string{
		"method": "HandleScheduleEvent",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, err)

	var activationData v1alpha2.ActivationData
	jData, _ := json.Marshal(event.Body)
	err = json.Unmarshal(jData, &activationData)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "event body is not a activation data", v1alpha2.BadRequest)
	}
	key := fmt.Sprintf("sch_%s-%s", activationData.Campaign, activationData.Activation)
	_, err = s.StateProvider.Upsert(ctx, states.UpsertRequest{
		Value: states.StateEntry{
			ID:   key,
			Body: activationData,
		},
	})
	return err
}
func (s *JobsManager) HandleJobEvent(ctx context.Context, event v1alpha2.Event) error {
	ctx, span := observability.StartSpan("Job Manager", ctx, &map[string]string{
		"method": "HandleJobEvent",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, err)

	if objectType, ok := event.Metadata["objectType"]; ok {
		var job v1alpha2.JobData
		jData, _ := json.Marshal(event.Body)
		err := json.Unmarshal(jData, &job)
		if err != nil {
			return v1alpha2.NewCOAError(nil, "event body is not a job", v1alpha2.BadRequest)
		}

		err = s.DelayOrSkipJob(ctx, objectType, job)
		if err != nil {
			return err
		}

		baseUrl, err := utils.GetString(s.Manager.Config.Properties, "baseUrl")
		if err != nil {
			return err
		}
		user, err := utils.GetString(s.Manager.Config.Properties, "user")
		if err != nil {
			return err
		}
		password, err := utils.GetString(s.Manager.Config.Properties, "password")
		if err != nil {
			return err
		}
		switch objectType {
		case "instance":
			instanceName := job.Id

			//get intance
			instance, err := utils.GetInstance(ctx, baseUrl, instanceName, user, password)
			if err != nil {
				return err //TODO: instance is gone
			}

			if instance.Status == nil {
				instance.Status = make(map[string]string)
			}

			//get solution
			solution, err := utils.GetSolution(ctx, baseUrl, instance.Spec.Solution, user, password)
			if err != nil {
				solution = model.SolutionState{
					Id: instance.Spec.Solution,
					Spec: &model.SolutionSpec{
						Components: make([]model.ComponentSpec, 0),
					},
				}
			}

			//get targets
			var targets []model.TargetState
			targets, err = utils.GetTargets(ctx, baseUrl, user, password)
			if err != nil {
				targets = make([]model.TargetState, 0)
			}

			//get target candidates
			targetCandidates := utils.MatchTargets(instance, targets)

			//create deployment spec
			deployment, err := utils.CreateSymphonyDeployment(instance, solution, targetCandidates, nil)
			if err != nil {
				return err
			}

			//call api
			if job.Action == "UPDATE" {
				_, err := utils.Reconcile(ctx, baseUrl, user, password, deployment, false)
				if err != nil {
					return err
				} else {
					s.StateProvider.Upsert(ctx, states.UpsertRequest{
						Value: states.StateEntry{
							ID: "i_" + instance.Id,
							Body: LastSuccessTime{
								Time: time.Now().UTC(),
							},
						},
					})
				}
			}
			if job.Action == "DELETE" {
				_, err := utils.Reconcile(ctx, baseUrl, user, password, deployment, true)
				if err != nil {
					return err
				} else {
					return utils.DeleteInstance(ctx, baseUrl, deployment.Instance.Name, user, password)
				}
			}
		case "target":
			targetName := job.Id
			target, err := utils.GetTarget(ctx, baseUrl, targetName, user, password)
			if err != nil {
				return err
			}
			deployment, err := utils.CreateSymphonyDeploymentFromTarget(target)
			if err != nil {
				return err
			}
			if job.Action == "UPDATE" {
				_, err := utils.Reconcile(ctx, baseUrl, user, password, deployment, false)
				if err != nil {
					return err
				} else {
					// TODO: how to handle status updates?
					s.StateProvider.Upsert(ctx, states.UpsertRequest{
						Value: states.StateEntry{
							ID: "t_" + targetName,
							Body: LastSuccessTime{
								Time: time.Now().UTC(),
							},
						},
					})
				}
			}
			if job.Action == "DELETE" {
				_, err := utils.Reconcile(ctx, baseUrl, user, password, deployment, true)
				if err != nil {
					return err
				} else {
					return utils.DeleteTarget(ctx, baseUrl, targetName, user, password)
				}
			}
		}
	}
	return nil
}
