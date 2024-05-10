/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package metahelper

import (
	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	MetaPopulator interface {
		PopulateMeta(object metaV1.Object, instance model.InstanceSpec) error
	}
	metaPopulator struct {
		annotationPopulators []func(instance model.InstanceSpec) (map[string]string, error)
		labelPopulators      []func(instance model.InstanceSpec) (map[string]string, error)
	}
	Option func(*metaPopulator)
)

func NewMetaPopulator(opts ...Option) (*metaPopulator, error) {
	mp := &metaPopulator{}
	for _, opt := range opts {
		opt(mp)
	}
	return mp, nil
}

func WithDefaultPopulators() Option {
	return func(mp *metaPopulator) {
		mp.annotationPopulators = append(mp.annotationPopulators, populateDefaultAnnotations)
		mp.labelPopulators = append(mp.labelPopulators, populateDefaultLabels)
	}
}

func (m *metaPopulator) PopulateMeta(object metaV1.Object, instance model.InstanceSpec) error {
	if err := m.populateLabels(object, instance); err != nil {
		return err
	}
	if err := m.populateAnnotations(object, instance); err != nil {
		return err
	}
	return nil
}

func (m *metaPopulator) populateLabels(object metaV1.Object, instance model.InstanceSpec) error {
	var labels []map[string]string
	labels = append(labels, object.GetLabels())
	for _, f := range m.labelPopulators {
		label, err := f(instance)
		if err != nil {
			return err
		}
		labels = append(labels, label)
	}
	object.SetLabels(utils.MergeCollection(labels...))
	return nil
}

func (m *metaPopulator) populateAnnotations(object metaV1.Object, instance model.InstanceSpec) error {
	var annotations []map[string]string
	annotations = append(annotations, object.GetAnnotations())
	for _, f := range m.annotationPopulators {
		annotation, err := f(instance)
		if err != nil {
			return err
		}
		annotations = append(annotations, annotation)
	}
	object.SetAnnotations(utils.MergeCollection(annotations...))
	return nil
}

func populateDefaultLabels(instance model.InstanceSpec) (map[string]string, error) {
	labels := make(map[string]string)
	labels[constants.ManagerMetaKey] = constants.API
	return labels, nil
}

func populateDefaultAnnotations(instance model.InstanceSpec) (map[string]string, error) {
	annotations := make(map[string]string)
	annotations[constants.InstanceMetaKey] = instance.Name
	return annotations, nil
}
