/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"context"
	"fmt"
)

// State represents a response state
type (
	State      uint16
	Terminable interface {
		Shutdown(ctx context.Context) error
	}
)

const (
	// OK = HTTP 200
	OK State = 200
	// Accepted = HTTP 202
	Accepted State = 202
	// BadRequest = HTTP 400
	BadRequest State = 400
	// Unauthorized = HTTP 403
	Unauthorized State = 403
	// NotFound = HTTP 404
	NotFound State = 404
	// MethodNotAllowed = HTTP 405
	MethodNotAllowed State = 405
	Conflict         State = 409
	// InternalError = HTTP 500
	InternalError State = 500
	// Config errors
	BadConfig     State = 1000
	MissingConfig State = 1001
	// API invocation errors
	InvalidArgument State = 2000
	APIRedirect     State = 3030
	// IO errors
	FileAccessError State = 4000
	// Serialization errors
	SerializationError State = 5000
	DeserializeError   State = 5001
	// Async requets
	DeleteRequested State = 6000
	// Operation results
	UpdateFailed   State = 8001
	DeleteFailed   State = 8002
	ValidateFailed State = 8003
	Updated        State = 8004
	Deleted        State = 8005
	// Workflow status
	Running        State = 9994
	Paused         State = 9995
	Done           State = 9996
	Delayed        State = 9997
	Untouched      State = 9998
	NotImplemented State = 9999

	// To have clearer metrics/self-explanatory errors, we introduce some
	// detailed error codes
	InitFailed                      State = 10000
	CreateActionConfigFailed        State = 10001
	HelmActionFailed                State = 10002
	GetComponentSpecFailed          State = 10003
	CreateProjectorFailed           State = 10004
	K8sRemoveServiceFailed          State = 10005
	K8sRemoveDeploymentFailed       State = 10006
	K8sDeploymentFailed             State = 10007
	ReadYamlFailed                  State = 10008
	ApplyYamlFailed                 State = 10009
	ReadResourcePropertyFailed      State = 10010
	ApplyResourceFailed             State = 10011
	DeleteYamlFailed                State = 10012
	DeleteResourceFailed            State = 10013
	CheckResourceStatusFailed       State = 10014
	ApplyScriptFailed               State = 10015
	RemoveScriptFailed              State = 10016
	YamlResourcePropertyNotFound    State = 10017
	GetHelmPropertyFailed           State = 10018
	HelmChartPullFailed             State = 10019
	HelmChartLoadFailed             State = 10020
	HelmChartApplyFailed            State = 10021
	HelmChartUninstallFailed        State = 10022
	IngressApplyFailed              State = 10023
	HttpNewRequestFailed            State = 10024
	HttpSendRequestFailed           State = 10025
	HttpErrorResponse               State = 10026
	MqttPublishFailed               State = 10027
	MqttApplyFailed                 State = 10028
	MqttApplyTimeout                State = 10029
	ConfigMapApplyFailed            State = 10030
	HttpBadWaitStatusCode           State = 10031
	HttpNewWaitRequestFailed        State = 10032
	HttpSendWaitRequestFailed       State = 10033
	HttpErrorWaitResponse           State = 10034
	HttpBadWaitExpression           State = 10035
	ScriptExecutionFailed           State = 10036
	ScriptResultParsingFailed       State = 10037
	WaitToGetInstancesFailed        State = 10038
	WaitToGetSitesFailed            State = 10039
	WaitToGetCatalogsFailed         State = 10040
	InvalidWaitObjectType           State = 10041
	CatalogsGetFailed               State = 10042
	InvalidInstanceCatalog          State = 10043
	CreateInstanceFromCatalogFailed State = 10044
	InvalidSolutionCatalog          State = 10045
	CreateSolutionFromCatalogFailed State = 10046
	InvalidTargetCatalog            State = 10047
	CreateTargetFromCatalogFailed   State = 10048
	InvalidCatalogCatalog           State = 10049
	CreateCatalogFromCatalogFailed  State = 10050
	ParentObjectMissing             State = 10051
	ParentObjectCreateFailed        State = 10052
	MaterializeBatchFailed          State = 10053

	// instance controller errors
	SolutionGetFailed             State = 11000
	TargetCandidatesNotFound      State = 11001
	TargetListGetFailed           State = 11002
	ObjectInstanceCoversionFailed State = 11003
	TimedOut                      State = 11004

	//target controller errors
	TargetPropertyNotFound State = 12000
)

func (s State) EqualsWithString(str string) bool {
	return s.String() == str
}

func (s State) String() string {
	switch s {
	case OK:
		return "OK"
	case Accepted:
		return "Accepted"
	case BadRequest:
		return "Bad Request"
	case Unauthorized:
		return "Unauthorized"
	case NotFound:
		return "Not Found"
	case MethodNotAllowed:
		return "Method Not Allowed"
	case Conflict:
		return "Conflict"
	case InternalError:
		return "Internal Error"
	case BadConfig:
		return "Bad Config"
	case MissingConfig:
		return "Missing Config"
	case InvalidArgument:
		return "Invalid Argument"
	case APIRedirect:
		return "API Redirect"
	case FileAccessError:
		return "File Access Error"
	case SerializationError:
		return "Serialization Error"
	case DeserializeError:
		return "De-serialization Error"
	case DeleteRequested:
		return "Delete Requested"
	case UpdateFailed:
		return "Update Failed"
	case DeleteFailed:
		return "Delete Failed"
	case ValidateFailed:
		return "Validate Failed"
	case Updated:
		return "Updated"
	case Deleted:
		return "Deleted"
	case Running:
		return "Running"
	case Paused:
		return "Paused"
	case Done:
		return "Done"
	case Delayed:
		return "Delayed"
	case Untouched:
		return "Untouched"
	case NotImplemented:
		return "Not Implemented"
	case InitFailed:
		return "Init Failed"
	case CreateActionConfigFailed:
		return "Create Action Config Failed"
	case HelmActionFailed:
		return "Helm Action Failed"
	case GetComponentSpecFailed:
		return "Get Component Spec Failed"
	case CreateProjectorFailed:
		return "Create Projector Failed"
	case K8sRemoveServiceFailed:
		return "Remove K8s Service Failed"
	case K8sRemoveDeploymentFailed:
		return "Remove K8s Deployment Failed"
	case K8sDeploymentFailed:
		return "K8s Deployment Failed"
	case ReadYamlFailed:
		return "Read Yaml Failed"
	case ApplyYamlFailed:
		return "Apply Yaml Failed"
	case ReadResourcePropertyFailed:
		return "Read Resource Property Failed"
	case ApplyResourceFailed:
		return "Apply Resource Failed"
	case DeleteYamlFailed:
		return "Delete Yaml Failed"
	case DeleteResourceFailed:
		return "Delete Resource Failed"
	case CheckResourceStatusFailed:
		return "Check Resource Status Failed"
	case ApplyScriptFailed:
		return "Apply Script Failed"
	case RemoveScriptFailed:
		return "Remove Script Failed"
	case YamlResourcePropertyNotFound:
		return "Yaml or Resource Property Not Found"
	case GetHelmPropertyFailed:
		return "Get Helm Property Failed"
	case HelmChartPullFailed:
		return "Helm Chart Pull Failed"
	case HelmChartLoadFailed:
		return "Helm Chart Load Failed"
	case HelmChartApplyFailed:
		return "Helm Chart Apply Failed"
	case HelmChartUninstallFailed:
		return "Helm Chart Uninstall Failed"
	case IngressApplyFailed:
		return "Ingress Apply Failed"
	case HttpNewRequestFailed:
		return "Http New Request Failed"
	case HttpSendRequestFailed:
		return "Http Send Request Failed"
	case HttpErrorResponse:
		return "Http Error Response"
	case MqttPublishFailed:
		return "Mqtt Publish Failed"
	case MqttApplyFailed:
		return "Mqtt Apply Failed"
	case MqttApplyTimeout:
		return "Mqtt Apply Timeout"
	case ConfigMapApplyFailed:
		return "ConfigMap Apply Failed"
	case HttpBadWaitStatusCode:
		return "Http Bad Wait Status Code"
	case HttpNewWaitRequestFailed:
		return "Http New Wait Request Failed"
	case HttpSendWaitRequestFailed:
		return "Http Send Wait Request Failed"
	case HttpErrorWaitResponse:
		return "Http Error Wait Response"
	case HttpBadWaitExpression:
		return "Http Bad Wait Expression"
	case ScriptExecutionFailed:
		return "Script Execution Failed"
	case ScriptResultParsingFailed:
		return "Script Result Parsing Failed"
	case WaitToGetInstancesFailed:
		return "Wait To Get Instances Failed"
	case WaitToGetSitesFailed:
		return "Wait To Get Sites Failed"
	case WaitToGetCatalogsFailed:
		return "Wait To Get Catalogs Failed"
	case InvalidWaitObjectType:
		return "Invalid Wait Object Type"
	case CatalogsGetFailed:
		return "Get Catalogs Failed"
	case InvalidInstanceCatalog:
		return "Invalid Instance Catalog"
	case CreateInstanceFromCatalogFailed:
		return "Create Instance From Catalog Failed"
	case InvalidSolutionCatalog:
		return "Invalid Solution Object in Catalog"
	case CreateSolutionFromCatalogFailed:
		return "Create Solution Object From Catalog Failed"
	case InvalidTargetCatalog:
		return "Invalid Target Object in Catalog"
	case CreateTargetFromCatalogFailed:
		return "Create Target Object From Catalog Failed"
	case InvalidCatalogCatalog:
		return "Invalid Catalog Object in Catalog"
	case CreateCatalogFromCatalogFailed:
		return "Create Catalog Object From Catalog Failed"
	case ParentObjectMissing:
		return "Parent Object Missing"
	case ParentObjectCreateFailed:
		return "Parent Object Create Failed"
	case MaterializeBatchFailed:
		return "Failed to Materialize all objects"
	case SolutionGetFailed:
		return "Solution does not exist"
	case TargetCandidatesNotFound:
		return "Target does not exist"
	case TargetListGetFailed:
		return "Target list does not exist"
	case ObjectInstanceCoversionFailed:
		return "Object to Instance conversion failed"
	case TimedOut:
		return "Timed Out"
	case TargetPropertyNotFound:
		return "Target Property Not Found"
	default:
		return fmt.Sprintf("Unknown State: %d", s)
	}
}

const (
	COAMetaHeader            = "COA_META_HEADER"
	TracingExporterConsole   = "tracing.exporters.console"
	MetricsExporterOTLPgRPC  = "metrics.exporters.otlpgrpc"
	TracingExporterZipkin    = "tracing.exporters.zipkin"
	TracingExporterOTLPgRPC  = "tracing.exporters.otlpgrpc"
	ProvidersPersistentState = "providers.persistentstate"
	ProvidersVolatileState   = "providers.volatilestate"
	ProvidersConfig          = "providers.config"
	ProvidersSecret          = "providers.secret"
	ProvidersReference       = "providers.reference"
	ProvidersProbe           = "providers.probe"
	ProvidersUploader        = "providers.uploader"
	ProvidersReporter        = "providers.reporter"
	ProviderQueue            = "providers.queue"
	ProviderLedger           = "providers.ledger"
	StatusOutput             = "__status"
	ErrorOutput              = "__error"
	StateOutput              = "__state"
)
