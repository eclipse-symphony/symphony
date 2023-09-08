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

package config

import (
	providers "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

type IConfigProvider interface {
	Init(config providers.IProviderConfig) error
	Read(object string, field string) (string, error)
	ReadObject(object string) (map[string]string, error)
	Set(object string, field string, value string) error
	SetObject(object string, value map[string]string) error
	Remove(object string, field string) error
	RemoveObject(object string) error
}

type IExtConfigProvider interface {
	Get(object string, field string, overrides []string) (string, error)
	GetObject(object string, overrides []string) (map[string]string, error)
}
