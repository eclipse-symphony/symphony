//go:build !arm && !arm64
// +build !arm,!arm64

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package rust

/*
#cgo LDFLAGS: -L./target/release -lrust_binding -lm -ldl -lpthread
*/
import "C"
