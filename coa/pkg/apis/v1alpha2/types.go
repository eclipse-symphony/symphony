/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"fmt"
)

// State represents a response state
type State uint16

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
)

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
	default:
		return fmt.Sprintf("Unknown State: %d", s)
	}
}

const (
	COAMetaHeader          = "COA_META_HEADER"
	TracingExporterConsole = "tracing.exporters.console"
	TracingExporterZipkin  = "tracing.exporters.zipkin"
	ProvidersState         = "providers.state"
	ProvidersConfig        = "providers.config"
	ProvidersSecret        = "providers.secret"
	ProvidersReference     = "providers.reference"
	ProvidersProbe         = "providers.probe"
	ProvidersUploader      = "providers.uploader"
	ProvidersReporter      = "providers.reporter"
	ProviderQueue          = "providers.queue"
	ProviderLedger         = "providers.ledger"
	StatusOutput           = "__status"
	ErrorOutput            = "__error"
	StateOutput            = "__state"
)
