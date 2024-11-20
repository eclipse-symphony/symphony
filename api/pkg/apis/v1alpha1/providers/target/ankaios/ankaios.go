/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package ankaios

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"

	ankbase "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/ankaios/ank_base"       // Replace with the actual path to the generated package
	controlapi "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/ankaios/control_api" // Replace with the actual path to the generated package
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const (
	ControlInterfaceBasePath = "/run/ankaios/control_interface"
	RequestID                = "dynamic_nginx@go_control_interface"
	WaitingTimeInSec         = 5
)

// Create the CompleteStateRequest
func createRequestForCompleteState() *controlapi.ToAnkaios {
	return &controlapi.ToAnkaios{
		ToAnkaiosEnum: &controlapi.ToAnkaios_Request{
			Request: &ankbase.Request{
				RequestId: RequestID,
				RequestContent: &ankbase.Request_CompleteStateRequest{
					CompleteStateRequest: &ankbase.CompleteStateRequest{
						FieldMask: []string{"workloadStates.agent_A.dynamic_nginx"},
					},
				},
			},
		},
	}
}

// Read Protobuf message from FIFO
func readProtobufData(file *os.File) ([]byte, error) {
	// Read the length of the upcoming message (varint encoded)
	var size uint64
	if err := binary.Read(file, binary.LittleEndian, &size); err != nil {
		return nil, err
	}

	// Read the actual message
	buf := make([]byte, size)
	_, err := file.Read(buf)
	return buf, err
}

// Read from Control Interface
func readFromControlInterface() {
	inputPath := ControlInterfaceBasePath + "/input"

	file, err := os.Open(inputPath)
	if err != nil {
		log.Fatalf("Failed to open input FIFO: %v", err)
	}
	defer file.Close()

	for {
		data, err := readProtobufData(file)
		if err != nil {
			log.Printf("Error reading data: %v", err)
			continue
		}

		var response controlapi.FromAnkaios
		if err := proto.Unmarshal(data, &response); err != nil {
			log.Printf("Failed to parse Protobuf data: %v", err)
			continue
		}

		if resp, ok := response.FromAnkaiosEnum.(*controlapi.FromAnkaios_Response); ok {
			log.Printf("Received Response: %v", resp.Response)
		} else {
			log.Println("Invalid response type")
		}
	}
}

// Write Protobuf message to FIFO
func writeProtobufData(file *os.File, message proto.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	// Write message length followed by the message itself
	if err := binary.Write(file, binary.LittleEndian, uint64(len(data))); err != nil {
		return err
	}
	_, err = file.Write(data)
	return err
}

// Write to Control Interface
func writeToControlInterface() {
	outputPath := ControlInterfaceBasePath + "/output"

	file, err := os.OpenFile(outputPath, os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open output FIFO: %v", err)
	}
	defer file.Close()

	// Send CompleteStateRequest
	for {
		request := createRequestForCompleteState()
		if err := writeProtobufData(file, request); err != nil {
			log.Printf("Failed to write request: %v", err)
		} else {
			log.Println("Sent CompleteStateRequest")
		}
		time.Sleep(WaitingTimeInSec * time.Second)
	}
}

func createRequestToAddNewWorkload() *controlapi.ToAnkaios {
	// Define the workload
	newWorkloads := &ankbase.WorkloadMap{
		Workloads: map[string]*ankbase.Workload{
			"dynamic_nginx": {
				Runtime: proto.String("podman"),
				Agent:   proto.String("agent_A"),
				RestartPolicy: func(policy ankbase.RestartPolicy) *ankbase.RestartPolicy {
					return &policy
				}(ankbase.RestartPolicy_NEVER),
				Tags: &ankbase.Tags{
					Tags: []*ankbase.Tag{
						{
							Key:   "owner",
							Value: "Ankaios team",
						},
					},
				},
				RuntimeConfig: proto.String("image: docker.io/library/nginx\ncommandOptions: [\"-p\", \"8080:80\"]"),
				Dependencies: &ankbase.Dependencies{
					Dependencies: map[string]ankbase.AddCondition{},
				},
				Configs:                nil,
				ControlInterfaceAccess: nil,
			},
		},
	}

	// Define the update state request
	updateStateRequest := &ankbase.UpdateStateRequest{
		NewState: &ankbase.CompleteState{
			DesiredState: &ankbase.State{
				ApiVersion: "v0.1",
				Workloads:  newWorkloads,
			},
		},
		UpdateMask: []string{"desiredState.workloads.dynamic_nginx"},
	}

	// Wrap it in a Request
	request := &ankbase.Request{
		RequestId: RequestID,
		RequestContent: &ankbase.Request_UpdateStateRequest{
			UpdateStateRequest: updateStateRequest,
		},
	}

	// Create the ToAnkaios message
	toAnkaios := &controlapi.ToAnkaios{
		ToAnkaiosEnum: &controlapi.ToAnkaios_Request{
			Request: request,
		},
	}

	return toAnkaios
}

type AnkaiosTargetProviderConfig struct {
	ID string `json:"id"`
}
type AnkaiosTargetProvider struct {
	Config  AnkaiosTargetProviderConfig
	Context *contexts.ManagerContext
}

var cache map[string][]model.ComponentSpec
var mLock sync.Mutex

func (m *AnkaiosTargetProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan(
		"Mock Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	mLock.Lock()
	defer mLock.Unlock()

	mockConfig, err := toMockTargetProviderConfig(config)
	if err != nil {
		return err
	}
	m.Config = mockConfig
	if cache == nil {
		cache = make(map[string][]model.ComponentSpec)
	}

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure()) // Replace with the server's address and use TLS if needed
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	go readFromControlInterface()
	writeToControlInterface()

	return nil
}
func (s *AnkaiosTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func toMockTargetProviderConfig(config providers.IProviderConfig) (AnkaiosTargetProviderConfig, error) {
	ret := AnkaiosTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *AnkaiosTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := MockTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func MockTargetProviderConfigFromMap(properties map[string]string) (AnkaiosTargetProviderConfig, error) {
	ret := AnkaiosTargetProviderConfig{}
	ret.ID = properties["id"]
	return ret, nil
}
func (m *AnkaiosTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	mLock.Lock()
	defer mLock.Unlock()

	_, span := observability.StartSpan(
		"Mock Target Provider",
		ctx,
		&map[string]string{
			"method": "Get",
		},
	)
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	ret := make([]model.ComponentSpec, 0)
	for _, c := range cache[m.Config.ID] {
		for _, r := range references {
			if c.Name == r.Component.Name {
				ret = append(ret, c)
				break
			}
		}
	}
	return ret, nil
}
func (m *AnkaiosTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	_, span := observability.StartSpan(
		"Mock Target Provider",
		ctx,
		&map[string]string{
			"method": "Apply",
		},
	)
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	mLock.Lock()
	defer mLock.Unlock()
	if cache[m.Config.ID] == nil {
		cache[m.Config.ID] = make([]model.ComponentSpec, 0)
	}
	for _, c := range step.Components {
		found := false
		for i, _ := range cache[m.Config.ID] {
			if cache[m.Config.ID][i].Name == c.Component.Name {
				found = true
				if c.Action == model.ComponentDelete {
					cache[m.Config.ID] = append(cache[m.Config.ID][:i], cache[m.Config.ID][i+1:]...)
				}
				break
			}
		}
		if !found {
			cache[m.Config.ID] = append(cache[m.Config.ID], c.Component)
		}
	}
	ret := make(map[string]model.ComponentResultSpec)
	for _, c := range cache[m.Config.ID] {
		ret[c.Name] = model.ComponentResultSpec{
			Status:  v1alpha2.OK,
			Message: "",
		}
	}
	return ret, nil
}
func (m *AnkaiosTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{}
}
