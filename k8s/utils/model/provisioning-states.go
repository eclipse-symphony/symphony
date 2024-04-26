/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import "sigs.k8s.io/controller-runtime/pkg/client"

// ARM has a strict requirement on what the terminal states need to be for the provisioning of resources through the LRO contract (Succeeded, Failed, Cancelled)
// The documentation that talks about this can be found here: https://armwiki.azurewebsites.net/rpaas/async.html#provisioningstate-property
// The below exported members capture these states. The first three are the terminal states required by ARM and the
// fourth is a non-terminal state we use to indicate that the resource is being reconciled.
type ProvisioningStatus string

const (
	ProvisioningStatusSucceeded   ProvisioningStatus = "Succeeded"
	ProvisioningStatusFailed      ProvisioningStatus = "Failed"
	ProvisioningStatusCancelled   ProvisioningStatus = "Cancelled"
	ProvisioningStatusReconciling ProvisioningStatus = "Reconciling"
	ProvisioningStatusDeleting    ProvisioningStatus = "Deleting"
)

func IsTerminalState(status string) bool {
	return status == string(ProvisioningStatusSucceeded) || status == string(ProvisioningStatusFailed)
}

func GetNonTerminalStatus(object client.Object) ProvisioningStatus {
	if object.GetDeletionTimestamp() != nil {
		return ProvisioningStatusDeleting
	}
	return ProvisioningStatusReconciling
}
