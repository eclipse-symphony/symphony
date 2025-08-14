/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"fmt"
)

type IRetriableError interface {
	IsRetriableErr() bool
}

type COAError struct {
	InnerError error
	Message    string
	State      State
}

func (e COAError) Error() string {
	if e.Message != "" && e.InnerError != nil {
		return fmt.Sprintf("%s: %s (caused by: %s)", e.State.String(), e.Message, e.InnerError.Error())
	} else if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.State.String(), e.Message)
	} else if e.InnerError != nil {
		return e.InnerError.Error()
	} else {
		return ""
	}
}

func (e COAError) IsUserErr() bool {
	// case BadRequest, Unauthorized, NotFound, BadConfig, MethodNotAllowed, Conflict, MissingConfig, InvalidArgument, DeserializeError, SerializationError:
	return e.State < 500 && e.State >= 400
}

func containsError(states []State, state State) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

func getNonRetriableManagerConfigErrors() []State {
	return []State{
		InitFailed, ValidateFailed, GetComponentPropsFailed,
	}
}

func getNonRetriableProviderConfigErrors() []State {
	return []State{
		CreateProjectorFailed,                           // k8s
		CreateActionConfigFailed, GetHelmPropertyFailed, // helm provider
	}
}

func (e COAError) IsRetriableErr() bool {
	if e.IsUserErr() {
		return false
	}
	if containsError(getNonRetriableManagerConfigErrors(), e.State) {
		return false
	}
	if containsError(getNonRetriableProviderConfigErrors(), e.State) {
		return false
	}

	// default:
	return true
}

func FromError(err error) COAError {
	return COAError{
		InnerError: err,
		Message:    err.Error(),
		State:      InternalError,
	}
}
func FromHTTPResponseCode(code int, body []byte) COAError {
	var state State
	switch code {
	case 400:
		state = BadRequest
	case 401:
		state = Unauthorized
	case 403:
		state = Forbidden
	case 404:
		state = NotFound
	case 405:
		state = MethodNotAllowed
	case 409:
		state = Conflict
	default:
		state = InternalError
	}
	return COAError{
		Message: string(body),
		State:   state,
	}
}

func GetErrorState(err error) State {
	if coaErr, ok := err.(COAError); ok {
		return coaErr.State
	}
	return InternalError
}

func NewCOAError(err error, msg string, state State) COAError {
	return COAError{
		InnerError: err,
		Message:    msg,
		State:      state,
	}
}
func IsNotFound(err error) bool {
	coaE, ok := err.(COAError)
	if !ok {
		return false
	}
	return coaE.State == NotFound
}

func IsCanceled(err error) bool {
	coaE, ok := err.(COAError)
	if !ok {
		return false
	}
	return coaE.State == Canceled
}

func IsDelayed(err error) bool {
	coaE, ok := err.(COAError)
	if !ok {
		return false
	}
	return coaE.State == Delayed
}
func IsBadConfig(err error) bool {
	coaE, ok := err.(COAError)
	if !ok {
		return false
	}
	return coaE.State == BadConfig
}

func IsRetriableErr(err error) bool {
	ret, ok := err.(IRetriableError)
	if !ok {
		return true
	}
	return ret.IsRetriableErr()
}
