//go:build windows
// +build windows

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package rust

/*
  #cgo LDFLAGS: -L./target/x86_64-pc-windows-gnu/release -lrust_binding -lm -lpthread -lws2_32 -ladvapi32 -lbcrypt -luserenv
*/
import "C"
