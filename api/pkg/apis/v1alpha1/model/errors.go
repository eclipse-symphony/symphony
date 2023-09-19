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

package model

// Defines an error in the ARM resource for long running operations
// +kubebuilder:object:generate=true
type ErrorType struct {
	Code    string        `json:"code,omitempty"`
	Message string        `json:"message,omitempty"`
	Target  string        `json:"target,omitempty"`
	Details []TargetError `json:"details,omitempty"`
}

// Defines an error for symphony target
// +kubebuilder:object:generate=true
type TargetError struct {
	Code    string           `json:"code,omitempty"`
	Message string           `json:"message,omitempty"`
	Target  string           `json:"target,omitempty"`
	Details []ComponentError `json:"details,omitempty"`
}

// Defines an error for components defined in symphony
// +kubebuilder:object:generate=true
type ComponentError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Target  string `json:"target,omitempty"`
}

// Defines the state of the ARM resource for long running operations
// +kubebuilder:object:generate=true
type ProvisioningStatus struct {
	OperationID  string            `json:"operationId"`
	Status       string            `json:"status"`
	FailureCause string            `json:"failureCause,omitempty"`
	LogErrors    bool              `json:"logErrors,omitempty"`
	Error        ErrorType         `json:"error,omitempty"`
	Output       map[string]string `json:"output,omitempty"`
}
