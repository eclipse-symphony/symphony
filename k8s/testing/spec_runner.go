/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package testing

import (
	"os"
	"testing"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
)

// While we still have a mix of Ginkgo Tests and regular Go tests, there are a
// few considerations to keep in mind:
//   - Running `go test` will run all tests, including Ginkgo tests. But it will
//     count each Ginkgo test as a single test, regardless of how many `It` specs
//     are in the test. Also when generating a report, it will include the
//     output of the ginkgo test including unicode control characters. This will
//     be inalid in the JUnit report.
//   - Running `ginkgo` will run both Ginkgo tests and regular Go tests. But it will
//     only generate a report for the Ginkgo tests. By default, the unicode control
//     characters will not be included in the report.
//
// Because of this, we need to disable unicode control characters for regular Go tests
// so that it's not included in the report. We would do so by checking the environment
// variable `GOUNIT` (which is set by `mage unitTest`) to determine if we should
// remove the unicode control characters
func RunGinkgoSpecs(t *testing.T, description string, args ...interface{}) bool {
	if _, ok := os.LookupEnv("GOUNIT"); ok {
		args = append(args, types.ReporterConfig{
			NoColor: true,
		})
	}
	return ginkgo.RunSpecs(t, description, args...)
}
