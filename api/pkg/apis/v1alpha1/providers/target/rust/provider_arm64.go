//go:build arm64
// +build arm64

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package rust

/*
#cgo LDFLAGS: -L./target/aarch64-unknown-linux-gnu/release -lrust_binding -lm -ldl -lpthread
*/
import "C"
