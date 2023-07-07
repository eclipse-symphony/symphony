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
	return s.Config.Properties["poll.enabled"] == "true"
}

func (s *JobsManager) Poll() []error {
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
	instances, err := utils.GetInstances(baseUrl, user, password)
	if err != nil {
		fmt.Println(err.Error())
		return []error{err}
	}
	for _, instance := range instances {
		entry, err := s.StateProvider.Get(context.Background(), states.GetRequest{
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
	targets, err := utils.GetTargets(baseUrl, user, password)
	if err != nil {
		fmt.Println(err.Error())
		return []error{err}
	}
	for _, target := range targets {
		entry, err := s.StateProvider.Get(context.Background(), states.GetRequest{
			ID: "t_" + target.Id,
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

func (s *JobsManager) Reconcil() []error {
	return nil
}
func (s *JobsManager) HandleHeartBeatEvent(ctx context.Context, event v1alpha2.Event) error {
	var heartbeat v1alpha2.HeartBeatData
	var jok bool
	if heartbeat, jok = event.Body.(v1alpha2.HeartBeatData); !jok {
		return v1alpha2.NewCOAError(nil, "event body is not a heart beat", v1alpha2.BadRequest)
	}
	// TODO: the heart beat data should contain a "finished" field so data can be cleared
	_, err := s.StateProvider.Upsert(ctx, states.UpsertRequest{
		Value: states.StateEntry{
			ID:   "h_" + heartbeat.JobId,
			Body: heartbeat,
		},
	})
	return err
}

func (s *JobsManager) DelayOrSkipJob(ctx context.Context, objectType string, job v1alpha2.JobData) error {
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

func (s *JobsManager) HandleJobEvent(ctx context.Context, event v1alpha2.Event) error {
	if objectType, ok := event.Metadata["objectType"]; ok {

		var job v1alpha2.JobData
		var jok bool
		if job, jok = event.Body.(v1alpha2.JobData); !jok {
			return v1alpha2.NewCOAError(nil, "event body is not a job", v1alpha2.BadRequest)
		}

		err := s.DelayOrSkipJob(ctx, objectType, job)
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
			instance, err := utils.GetInstance(baseUrl, instanceName, user, password)
			if err != nil {
				return err //TODO: instance is gone
			}

			if instance.Status == nil {
				instance.Status = make(map[string]string)
			}

			//get solution
			solution, err := utils.GetSolution(baseUrl, instance.Spec.Solution, user, password)
			if err != nil {
				// TODO: how to handle status updates?
				// instance.Status["status"] = "Solution Missing"
				// instance.Status["status-details"] = fmt.Sprintf("unable to fetch Solution object: %v", err)
				solution = model.SolutionState{
					Id: instance.Spec.Solution,
					Spec: &model.SolutionSpec{
						Components: make([]model.ComponentSpec, 0),
					},
				}
			}

			//get targets
			var targets []model.TargetState
			targets, err = utils.GetTargets(baseUrl, user, password)
			if err != nil {
				// TODO: how to handle status updates?
				// instance.Status["status"] = "No Targets"
				// instance.Status["status-details"] = fmt.Sprintf("unable to fetch Target objects: %v", err)
				targets = make([]model.TargetState, 0)
			}

			//get target candidates
			targetCandidates := utils.MatchTargets(instance, targets)
			if len(targetCandidates) == 0 {
				// TODO: how to handle status updates?
				// instance.Status["status"] = "No Matching Targets"
				// instance.Status["status-details"] = "no Targets are selected"
				return err //TODO: no match targets
			}

			//create deployment spec
			deployment, err := utils.CreateSymphonyDeployment(instance, solution, targetCandidates, nil)
			if err != nil {
				// TODO: how to handle status updates?
				// instance.Status["status"] = "Creation failed"
				// instance.Status["status-details"] = fmt.Sprintf("failed to generate Symphony deployment: %v", err)
			}

			//call api
			if job.Action == "UPDATE" {
				_, err := utils.Reconcile(baseUrl, user, password, deployment, false)
				if err != nil {
					// TODO: how to handle status updates?
					// instance.Status["status"] = "Failed"
					// instance.Status["status-details"] = fmt.Sprintf("failed to deploy: %v", summary)
				} else {
					// TODO: how to handle status updates?
					// instance.Status["status"] = "OK"
					// instance.Status["status-details"] = fmt.Sprintf("deployment summary: %v", summary)
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
				_, err := utils.Reconcile(baseUrl, user, password, deployment, true)
				if err != nil {
					// TODO: how to handle status updates?
					// instance.Status["status"] = "Remove Failed"
					// instance.Status["status-details"] = fmt.Sprintf("failed to remove: %v", summary)
				} else {
					// Instance is gone!
					return utils.DeleteInstance(baseUrl, deployment.Instance.Name, user, password)
				}
			}
		case "target":
			targetName := job.Id
			target, err := utils.GetTarget(baseUrl, targetName, user, password)
			if err != nil {
				return err
			}
			deployment, err := utils.CreateSymphonyDeploymentFromTarget(target)
			if err != nil {
				return err
			}
			if job.Action == "UPDATE" {
				_, err := utils.Reconcile(baseUrl, user, password, deployment, false)
				if err != nil {
					// TODO: how to handle status updates?
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
				_, err := utils.Reconcile(baseUrl, user, password, deployment, true)
				if err != nil {
					// TODO: how to handle status updates?
				} else {
					// Instance is gone!
					return utils.DeleteTarget(baseUrl, targetName, user, password)
				}
			}
		}
	}
	return nil
}
