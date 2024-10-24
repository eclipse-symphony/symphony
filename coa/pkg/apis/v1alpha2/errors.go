/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import "fmt"

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
	case 403:
		state = Unauthorized
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
	coaE, ok := err.(COAError)
	if !ok {
		return true
	}
	switch coaE.State {
	case BadRequest, Unauthorized, NotFound, BadConfig, MethodNotAllowed, Conflict, MissingConfig, InvalidArgument, DeserializeError, SerializationError:
		return false
	case ValidateFailed: // catalog manager
		return false
	default:
		return true
	}
}
