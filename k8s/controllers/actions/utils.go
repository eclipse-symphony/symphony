/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package actions

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CheckNeedtoDelete(meta metav1.ObjectMeta) bool {
	return meta.CreationTimestamp.Add(10 * time.Hour).Before(time.Now())
}
