/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package ledger

import (
	"context"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
)

type ILedgerProvider interface {
	Append(ctx context.Context, entries []v1alpha2.Trail) error
}
