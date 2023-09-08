//go:build mage

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

package main

import (
	"os"

	//mage:import
	_ "github.com/azure/symphony/packages/mage"
	"github.com/princjef/mageutil/shellcmd"
)

func Build() error {
	return shellcmd.Command("go build -o bin/symphony-api").Run()
}
func BuildAzure() error {
	return shellcmd.Command("go build -o bin/symphony-api -tags=azure").Run()
}

// Runs both api unit tests as well as coa unit tests.
func TestWithCoa() error {
	// Unit tests for api
	testHelper()

	// Change directory to coa
	os.Chdir("../coa")

	// Unit tests for coa
	testHelper()
	return nil
}

func testHelper() error {
	if err := shellcmd.RunAll(
		"go clean -testcache",
		"go test -race -timeout 30s -cover ./...",
	); err != nil {
		return err
	}
	return nil
}
