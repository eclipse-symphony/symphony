/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

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
