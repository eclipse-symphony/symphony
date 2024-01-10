/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8s

import (
	"context"
	"encoding/json"
	"path/filepath"

	"strconv"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type K8sReferenceProviderConfig struct {
	Name       string `json:"name"`
	ConfigPath string `json:"configPath"`
	InCluster  bool   `json:"inCluster"` //TODO: add context support
}

func K8sReferenceProviderConfigFromMap(properties map[string]string) (K8sReferenceProviderConfig, error) {
	ret := K8sReferenceProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	if v, ok := properties["configPath"]; ok {
		ret.ConfigPath = utils.ParseProperty(v)
	}
	if v, ok := properties["inCluster"]; ok {
		val := utils.ParseProperty(v)
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'inCluster' setting of K8s reference provider", v1alpha2.BadConfig)
			}
			ret.InCluster = bVal
		}
	}
	return ret, nil
}

func (i *K8sReferenceProvider) InitWithMap(properties map[string]string) error {
	config, err := K8sReferenceProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

type K8sReferenceProvider struct {
	Config        K8sReferenceProviderConfig
	Client        *kubernetes.Clientset
	DynamicClient dynamic.Interface
	Context       *contexts.ManagerContext
}

func (m *K8sReferenceProvider) ID() string {
	return m.Config.Name
}

func (m *K8sReferenceProvider) TargetID() string {
	return "DON'T USE"
}

func (m *K8sReferenceProvider) ReferenceType() string {
	return "v1alpha2.ReferenceK8sCRD"
}

func (m *K8sReferenceProvider) Reconfigure(config providers.IProviderConfig) error {
	return nil
}

func (a *K8sReferenceProvider) SetContext(context *contexts.ManagerContext) {
	a.Context = context
}

func (m *K8sReferenceProvider) Init(config providers.IProviderConfig) error {
	var err error
	aConfig, err := toK8sReferenceProviderConfig(config)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "provided config is not a valid mock config provider config", v1alpha2.BadConfig)
	}
	m.Config = aConfig
	var kConfig *rest.Config

	if m.Config.InCluster {
		kConfig, err = rest.InClusterConfig()
	} else {
		if m.Config.ConfigPath == "" {
			if home := homedir.HomeDir(); home != "" {
				m.Config.ConfigPath = filepath.Join(home, ".kube", "config")
			} else {
				return v1alpha2.NewCOAError(nil, "can't locate home direction to read default kubernetes config file, to run in cluster, set inCluster config setting to true", v1alpha2.BadConfig)
			}
		}
		kConfig, err = clientcmd.BuildConfigFromFlags("", m.Config.ConfigPath)
	}
	if err != nil {
		return err
	}
	m.Client, err = kubernetes.NewForConfig(kConfig)
	if err != nil {
		return err
	}
	m.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		return err
	}
	return nil
}

func toK8sReferenceProviderConfig(config providers.IProviderConfig) (K8sReferenceProviderConfig, error) {
	ret := K8sReferenceProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	ret.Name = utils.ParseProperty(ret.Name)
	ret.ConfigPath = utils.ParseProperty(ret.ConfigPath)
	return ret, err
}

func (m *K8sReferenceProvider) Get(id string, namespace string, group string, kind string, version string, ref string) (interface{}, error) {
	obj, err := m.DynamicClient.Resource(schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: kind,
	}).Namespace(namespace).Get(context.TODO(), id, v1.GetOptions{})

	if err != nil {
		return nil, err
	}
	//return obj.Object["spec"], nil
	return obj, nil
}

func (m *K8sReferenceProvider) List(labelSelector string, fieldSelector string, namespace string, group string, kind string, version string, ref string) (interface{}, error) {
	obj, err := m.DynamicClient.Resource(schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: kind,
	}).Namespace(namespace).List(context.TODO(), v1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	})

	if err != nil {
		return nil, err
	}

	ret := make([]interface{}, 0)
	for _, i := range obj.Items {
		//ret = append(ret, i.Object["spec"])
		ret = append(ret, i)
	}

	return ret, nil
}

func (a *K8sReferenceProvider) Clone(config providers.IProviderConfig) (providers.IProvider, error) {
	ret := &K8sReferenceProvider{}
	if config == nil {
		err := ret.Init(a.Config)
		if err != nil {
			return nil, err
		}
	} else {
		err := ret.Init(config)
		if err != nil {
			return nil, err
		}
	}
	if a.Context != nil {
		ret.Context = a.Context
	}
	return ret, nil
}
