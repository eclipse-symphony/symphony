//go:build remote

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

	// if globalLoggers only contains one logger, return this logger
	if len(globalLoggers) == 1 {
		for _, logger := range globalLoggers {
			return logger
		}
	} else if len(globalLoggers) < 1 {
		logger := newFileLogger(name, hooks.ContextHookOptions{DiagnosticLogContextEnabled: true, ActivityLogContextEnabled: false, Folding: true})
		globalLoggers[name] = logger
		return logger
	}
	panic(1)
}
