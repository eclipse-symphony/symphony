/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package azure

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetTokenNoSecret(t *testing.T) {
	tenantId := os.Getenv("TEST_AZURE_TENANT_ID")
	clientId := os.Getenv("TEST_AZURE_CLIENT_ID")
	if tenantId == "" || clientId == "" {
		t.Skip("Skipping because TEST_AZURE_TENANT_ID or TEST_AZURE_CLIENT_ID enviornment variable is not set")
	}
	token, err := GetAzureToken(tenantId, clientId, "", "https://api.adu.microsoft.com/.default")
	assert.Nil(t, err)
	assert.Equal(t, "", token)
}

func TestGetADUGroup(t *testing.T) {
	tenantId := os.Getenv("TEST_AZURE_TENANT_ID")
	clientId := os.Getenv("TEST_AZURE_CLIENT_ID")
	clientSecret := os.Getenv("TEST_AZURE_CLIENT_SECRET")
	aduAccountEndpoint := os.Getenv("TEST_ADU_ACCOUNT_ENDPOINT")
	aduAccountInstance := os.Getenv("TEST_ADU_ACCOUNT_INSTANCE")
	aduGroup := os.Getenv("TEST_ADU_GROUP")

	if tenantId == "" || clientId == "" || clientSecret == "" || aduAccountEndpoint == "" || aduAccountInstance == "" || aduGroup == "" {
		t.Skip("Skipping because TEST_AZURE_TENANT_ID, TEST_AZURE_CLIENT_ID, TEST_AZURE_CLIENT_SECRET, TEST_ADU_ACCOUNT_ENDPOINT, TEST_ADU_ACCOUNT_INSTANCE, or TEST_ADU_GROUP enviornment variable is not set")
	}

	token, err := GetAzureToken(tenantId, clientId, clientSecret, "https://api.adu.microsoft.com/.default")
	assert.Nil(t, err)
	assert.NotEqual(t, "", token)
	s, err := GetADUGroup(token, aduAccountEndpoint, aduAccountInstance, aduGroup)
	assert.NotNil(t, err)
	assert.Equal(t, aduGroup, s.GroupId)
}

func TestGetADUDeployment(t *testing.T) {
	tenantId := os.Getenv("TEST_AZURE_TENANT_ID")
	clientId := os.Getenv("TEST_AZURE_CLIENT_ID")
	clientSecret := os.Getenv("TEST_AZURE_CLIENT_SECRET")
	aduAccountEndpoint := os.Getenv("TEST_ADU_ACCOUNT_ENDPOINT")
	aduAccountInstance := os.Getenv("TEST_ADU_ACCOUNT_INSTANCE")
	aduGroup := os.Getenv("TEST_ADU_GROUP")
	aduDeloyment := os.Getenv("TEST_ADU_DEPLOYMENT")
	aduUpdateProvider := os.Getenv("TEST_ADU_UPDATE_PROVIDER")

	if tenantId == "" || clientId == "" || clientSecret == "" || aduAccountEndpoint == "" || aduAccountInstance == "" || aduGroup == "" || aduDeloyment == "" || aduUpdateProvider == "" {
		t.Skip("Skipping because TEST_AZURE_TENANT_ID, TEST_AZURE_CLIENT_ID, TEST_AZURE_CLIENT_SECRET, TEST_ADU_ACCOUNT_ENDPOINT, TEST_ADU_ACCOUNT_INSTANCE, TEST_ADU_GROUP, TEST_ADU_DEPLOYMENT or TEST_ADU_UPDATE_PROVIDER enviornment variable is not set")
	}
	token, err := GetAzureToken(tenantId, clientId, clientSecret, "https://api.adu.microsoft.com/.default")
	assert.Nil(t, err)
	assert.NotEqual(t, "", token)
	s, err := GetADUDeployment(token, aduAccountEndpoint, aduAccountInstance, aduGroup, aduDeloyment)
	assert.NotNil(t, err)
	assert.Equal(t, aduUpdateProvider, s.UpdateId.Provider)
}

func TestUpdateADUDeployment(t *testing.T) {
	tenantId := os.Getenv("TEST_AZURE_TENANT_ID")
	clientId := os.Getenv("TEST_AZURE_CLIENT_ID")
	clientSecret := os.Getenv("TEST_AZURE_CLIENT_SECRET")
	aduAccountEndpoint := os.Getenv("TEST_ADU_ACCOUNT_ENDPOINT")
	aduAccountInstance := os.Getenv("TEST_ADU_ACCOUNT_INSTANCE")
	aduGroup := os.Getenv("TEST_ADU_GROUP")
	aduDeloyment := os.Getenv("TEST_ADU_DEPLOYMENT")
	aduUpdateProvider := os.Getenv("TEST_ADU_UPDATE_PROVIDER")

	if tenantId == "" || clientId == "" || clientSecret == "" || aduAccountEndpoint == "" || aduAccountInstance == "" || aduGroup == "" || aduDeloyment == "" || aduUpdateProvider == "" {
		t.Skip("Skipping because TEST_AZURE_TENANT_ID, TEST_AZURE_CLIENT_ID, TEST_AZURE_CLIENT_SECRET, TEST_ADU_ACCOUNT_ENDPOINT, TEST_ADU_ACCOUNT_INSTANCE, TEST_ADU_GROUP, TEST_ADU_DEPLOYMENT or TEST_ADU_UPDATE_PROVIDER enviornment variable is not set")
	}

	token, err := GetAzureToken(tenantId, clientId, clientSecret, "https://api.adu.microsoft.com/.default")
	assert.Nil(t, err)
	assert.NotEqual(t, "", token)
	group, err := GetADUGroup(token, aduAccountEndpoint, aduAccountInstance, aduGroup)
	assert.NotNil(t, err)

	targetVersion := "1.0.0"

	if group.DeploymentId == "" {
		id := uuid.New().String()
		deployment := ADUDeployment{
			GroupId:      aduGroup,
			DeploymentId: id,
			UpdateId: UpdateId{
				Provider: aduUpdateProvider,
				Version:  targetVersion,
				Name:     "SAME54", //TODO: is this okay, or needs to be moved to environment varialbes as well?
			},
			StartDateTime: time.Now().UTC().Format("2006-01-02T15:04:05-0700"),
		}
		err = CreateADUDeployment(token, aduAccountEndpoint, aduAccountInstance, aduGroup, id, deployment)
		assert.NotNil(t, err)
	} else {
		s, err := GetADUDeployment(token, aduAccountEndpoint, aduAccountInstance, aduGroup, group.DeploymentId)
		if err != nil || s.UpdateId.Version != targetVersion {
			id := uuid.New().String()
			deployment := ADUDeployment{
				GroupId:      aduGroup,
				DeploymentId: id,
				UpdateId: UpdateId{
					Provider: aduUpdateProvider,
					Version:  targetVersion,
					Name:     "SAME54",
				},
				StartDateTime: time.Now().UTC().Format("2006-01-02T15:04:05-0700"),
			}

			err = CreateADUDeployment(token, aduAccountEndpoint, aduAccountInstance, aduGroup, id, deployment)
			assert.Nil(t, err)
		} else {
			if s.IsCanceled {
				s.StartDateTime = time.Now().UTC().Format("2006-01-02T15:04:05-0700")
				err = RetryADUDeployment(token, aduAccountEndpoint, aduAccountInstance, aduGroup, group.DeploymentId, s) //TODO: is group.DeploymntId parameter correct here?
				assert.Nil(t, err)
			}
		}
	}
}

func TestCreateResourceGroup(t *testing.T) {
	tenantId := os.Getenv("TEST_AZURE_TENANT_ID")
	clientId := os.Getenv("TEST_AZURE_CLIENT_ID")
	clientSecret := os.Getenv("TEST_AZURE_CLIENT_SECRET")
	resourceGroup := os.Getenv("TEST_AZURE_RESOURCE_GROUP")
	location := os.Getenv("TEST_AZURE_LOCATION")
	subscription := os.Getenv("TEST_AZURE_SUBSCRIPTION_ID")

	if tenantId == "" || clientId == "" || clientSecret == "" || resourceGroup == "" || location == "" || subscription == "" {
		t.Skip("Skipping because TEST_AZURE_TENANT_ID, TEST_AZURE_CLIENT_ID, TEST_AZURE_CLIENT_SECRET, TEST_AZURE_RESOURCE_GROUP, TEST_AZURE_SUBSCRIPTION_ID, or TEST_AZURE_LOCATION enviornment variable is not set")
	}

	token, err := GetAzureToken(tenantId, clientId, clientSecret, "https://management.azure.com/.default")
	assert.Nil(t, err)
	assert.NotEqual(t, "", token)
	err = CreateResourceGroup(token, subscription, resourceGroup, location)
	assert.Nil(t, err)
}
