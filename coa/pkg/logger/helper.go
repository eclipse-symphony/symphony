//go:build !remote

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package logger

import (
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/hooks"
)

// NewLogger creates new Logger instance.
func NewLogger(name string) Logger {
	globalLoggersLock.Lock()
	defer globalLoggersLock.Unlock()

	logger, ok := globalLoggers[name]
	if !ok {
		logger = newCoaLogger(name, hooks.ContextHookOptions{DiagnosticLogContextEnabled: true, ActivityLogContextEnabled: false, Folding: true})
		globalLoggers[name] = logger
	}

	return logger
}
