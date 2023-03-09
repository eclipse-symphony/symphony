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
	InternalError      State = 500
	BadConfig          State = 1000
	MissingConfig      State = 1001
	InvalidArgument    State = 2000
	APIRedirect        State = 3030
	FileAccessError    State = 4000
	SerializationError State = 5000
	Untouched          State = 9998
	NotImplemented     State = 9999
)

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
)
