/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package provisioningstates

// ARM has a strict requirement on what the terminal states need to be for the provisioning of resources through the LRO contract (Succeeded, Failed, Cancelled)
// The documentation that talks about this can be found here: https://armwiki.azurewebsites.net/rpaas/async.html#provisioningstate-property
// The below exported members capture these states. The first three are the terminal states required by ARM and the
// fourth is a non-terminal state we use to indicate that the resource is being reconciled.
const (
	Succeeded   = "Succeeded"
	Failed      = "Failed"
	Cancelled   = "Cancelled"
	Reconciling = "Reconciling"
)
