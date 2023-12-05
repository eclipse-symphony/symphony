/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

type COAError struct {
	InnerError error
	Message    string
	State      State
}

func (e COAError) Error() string {
	ret := e.Message
	if e.InnerError != nil {
		ret += " (" + e.InnerError.Error() + ")"
	}
	return ret
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
