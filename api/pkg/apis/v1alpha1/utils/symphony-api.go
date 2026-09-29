/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var (
	SymphonyAPIAddressBase = "http://symphony-service:8080/v1alpha2/"
	apiCertPath            = os.Getenv(constants.ApiCertEnvName)
)

func GetSymphonyAPIAddressBase() string {
	if os.Getenv(constants.SymphonyAPIUrlEnvName) == "" {
		return SymphonyAPIAddressBase
	}
	return os.Getenv(constants.SymphonyAPIUrlEnvName)
}

var symphonyApiClients sync.Map

func GetApiClient() (*apiClient, error) {
	symphonyBaseUrl := os.Getenv(constants.SymphonyAPIUrlEnvName)
	if value, ok := symphonyApiClients.Load(symphonyBaseUrl); ok {
		client, ok := value.(*apiClient)
		if !ok {
			log.Info("Symphony base url apiclient is broken. Recreating it.")
		} else {
			return client, nil
		}
	}
	log.Info("Creating the symphony base url apiclient.")
	client, err := getApiClient()
	if err != nil {
		log.Errorf("Failed to create the apiclient: %+v", err.Error())
		return nil, err
	}
	symphonyApiClients.Store(symphonyBaseUrl, client)
	return client, nil
}

func getApiClient() (*apiClient, error) {
	clientOptions := make([]ApiClientOption, 0)
	baseUrl := GetSymphonyAPIAddressBase()
	if caCert, ok := os.LookupEnv(constants.ApiCertEnvName); ok {
		clientOptions = append(clientOptions, WithCertAuth(caCert))
	}

	if ShouldUseSATokens() {
		log.Infof("Configuring API client with service account token provider")
		clientOptions = append(clientOptions, WithServiceAccountToken())
	} else {
		log.Infof("Configuring API client with user/password token provider")
		clientOptions = append(clientOptions, WithUserPassword(context.TODO()))
	}

	client, err := NewApiClient(context.Background(), baseUrl, clientOptions...)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetParentApiClient(baseUrl string) (*apiClient, error) {
	clientOptions := make([]ApiClientOption, 0)

	if caCert, ok := os.LookupEnv(constants.ApiCertEnvName); ok {
		clientOptions = append(clientOptions, WithCertAuth(caCert))
	}

	log.Infof("Configuring parent API client with user/password token provider for baseUrl: %s", baseUrl)
	clientOptions = append(clientOptions, WithUserPassword(context.TODO()))
	client, err := NewApiClient(context.Background(), baseUrl, clientOptions...)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func ShouldUseSATokens() bool {
	raw, ok := os.LookupEnv(constants.UseServiceAccountTokenEnvName)
	if !ok {
		return true
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return true
	}

	v, err := strconv.ParseBool(strings.ToLower(raw))
	if err != nil {
		log.Warnf("Invalid value for %s: %q; defaulting to service account token auth", constants.UseServiceAccountTokenEnvName, raw)
		return true
	}
	return v
}

func ShouldUseUserCreds() bool {
	return !ShouldUseSATokens()
}

var log = logger.NewLogger("coa.runtime")
