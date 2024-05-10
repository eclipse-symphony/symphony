/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"fmt"
	"gopls-workspace/apis/metrics/v1"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var devicelog = logf.Log.WithName("device-resource")
var myDeviceClient client.Client
var deviceWebhookValidationMetrics *metrics.Metrics

func (r *Device) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myDeviceClient = mgr.GetClient()

	mgr.GetFieldIndexer().IndexField(context.Background(), &Device{}, ".spec.displayName", func(rawObj client.Object) []string {
		device := rawObj.(*Device)
		return []string{device.Spec.DisplayName}
	})

	// initialize the controller operation metrics
	if deviceWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		deviceWebhookValidationMetrics = metrics
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

var _ webhook.Defaulter = &Device{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Device) Default() {
	devicelog.Info("default", "name", r.Name)

	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

var _ webhook.Validator = &Device{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Device) ValidateCreate() error {
	devicelog.Info("validate create", "name", r.Name)

	validateCreateTime := time.Now()
	validationError := r.validateCreateDevice()

	if validationError != nil {
		deviceWebhookValidationMetrics.ControllerValidationLatency(
			validateCreateTime,
			metrics.CreateOperationType,
			metrics.InvalidResource,
			metrics.DeviceResourceType,
		)
	} else {
		deviceWebhookValidationMetrics.ControllerValidationLatency(
			validateCreateTime,
			metrics.CreateOperationType,
			metrics.ValidResource,
			metrics.DeviceResourceType,
		)
	}

	return validationError
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Device) ValidateUpdate(old runtime.Object) error {
	devicelog.Info("validate update", "name", r.Name)

	validateUpdateTime := time.Now()
	validationError := r.validateUpdateDevice()

	if validationError != nil {
		deviceWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.InvalidResource,
			metrics.DeviceResourceType,
		)
	} else {
		deviceWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.ValidResource,
			metrics.DeviceResourceType,
		)
	}

	return validationError
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Device) ValidateDelete() error {
	devicelog.Info("validate delete", "name", r.Name)

	return nil
}

func (r *Device) validateCreateDevice() error {
	var devices DeviceList
	myDeviceClient.List(context.Background(), &devices, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if len(devices.Items) != 0 {
		return fmt.Errorf("device display name '%s' is already taken", r.Spec.DisplayName)
	}
	return nil
}

func (r *Device) validateUpdateDevice() error {
	var devices DeviceList
	err := myDeviceClient.List(context.Background(), &devices, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if err != nil {
		return err
	}
	if !(len(devices.Items) == 0 || len(devices.Items) == 1 && devices.Items[0].ObjectMeta.Name == r.ObjectMeta.Name) {
		return fmt.Errorf("device display name '%s' is already taken", r.Spec.DisplayName)
	}
	return nil
}
