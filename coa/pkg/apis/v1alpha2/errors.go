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
