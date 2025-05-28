//go:build arm
// +build arm

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package rust

/*
 #cgo LDFLAGS: -L./target/armv7-unknown-linux-gnueabihf/release -lsymphony -lm -ldl -lpthread
*/
import "C"
