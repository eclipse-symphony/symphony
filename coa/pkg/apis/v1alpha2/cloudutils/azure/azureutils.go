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
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
)

type AzureToken struct {
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    uint64 `json:"expires_in,omitempty"`
	ExtExpiresIn uint64 `json:"ext_expires_in,omitempty"`
	AccessToken  string `json:"access_token"`
}

type UpdateId struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
	Version  string `json:"version"`
}

type ADUDeployment struct {
	UpdateId      UpdateId `json:"updateId"`
	IsCanceled    bool     `json:"isCanceled,omitempty"`
	IsRetry       bool     `json:"isRetry,omitempty"`
	GroupId       string   `json:"groupId"`
	DeploymentId  string   `json:"deploymentId"`
	StartDateTime string   `json:"startDateTime,omitempty"`
}

type ADUGroup struct {
	GroupId         string   `json:"groupId"`
	GroupType       string   `json:"groupType"`
	Tags            []string `json:"tags,omitempty"`
	CreatedDateTime string   `json:"createdDateTime"`
	DeviceClassId   string   `json:"deviceClassId"`
	DeviceCount     uint64   `json:"deviceCount"`
	DeploymentId    string   `json:"deploymentId"`
}

func CreateSASToken(resourceUri string, keyName string, key string) string {
	span := time.Now().UTC().Sub(time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC))
	week := 60 * 60 * 27 * 7
	expiry := strconv.Itoa(int(span.Seconds()) + week)
	stringToSign := url.QueryEscape(resourceUri) + "\n" + expiry
	keyBytes, _ := base64.StdEncoding.DecodeString(key)
	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(stringToSign))
	signatureString := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	sasToken := fmt.Sprintf("SharedAccessSignature sr=%s&sig=%s&se=%s&skn=%s",
		url.QueryEscape(resourceUri),
		url.QueryEscape(signatureString),
		expiry,
		keyName)
	return sasToken
}

func GetAzureToken(tenantId string, clientId string, clientSecret string, scope string) (string, error) {
	authUrl := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantId)
	values := url.Values{}
	values.Add("client_id", clientId)
	values.Add("grant_type", "client_credentials")
	values.Add("scope", scope)
	values.Add("client_secret", clientSecret)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, authUrl, strings.NewReader(values.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var token AzureToken
	err = json.Unmarshal(result, &token)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func callRESTAPI(method string, getUrl string, token string, body io.Reader) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, getUrl, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	//req.Header.Set("Host", "microsoftonline.com")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to set HTTP request: %s", string(bodyBytes)), v1alpha2.InternalError)
	}

	return bodyBytes, nil
}

func GetADUGroup(token string, aduAccountEndpoint string, aduAccountInstance string, group string) (ADUGroup, error) {
	retGroup := ADUGroup{}
	getUrl := fmt.Sprintf("https://%s/deviceupdate/%s/management/groups/%s?api-version=2021-06-01-preview", aduAccountEndpoint, aduAccountInstance, group)
	ret, err := callRESTAPI("GET", getUrl, token, nil)
	if err != nil {
		return retGroup, err
	}
	json.Unmarshal(ret, &retGroup)
	if err != nil {
		return retGroup, err
	}
	return retGroup, nil
}

func GetADUDeployment(token string, aduAccountEndpoint string, aduAccountInstance string, group string, deployment string) (ADUDeployment, error) {
	retDeployment := ADUDeployment{}
	getUrl := fmt.Sprintf("https://%s/deviceupdate/%s/management/groups/%s/deployments/%s?api-version=2021-06-01-preview", aduAccountEndpoint, aduAccountInstance, group, deployment)
	ret, err := callRESTAPI("GET", getUrl, token, nil)
	if err != nil {
		return retDeployment, err
	}
	json.Unmarshal(ret, &retDeployment)
	if err != nil {
		return retDeployment, err
	}
	return retDeployment, nil
}

func RetryADUDeployment(token string, aduAccountEndpoint string, aduAccountInstance string, group string, deployment string, aduDeployment ADUDeployment) error {
	getUrl := fmt.Sprintf("https://%s/deviceupdate/%s/management/groups/%s/deployments/%s?action=retry&api-version=2021-06-01-preview", aduAccountEndpoint, aduAccountInstance, group, deployment)
	data, _ := json.Marshal(aduDeployment)
	_, err := callRESTAPI("POST", getUrl, token, bytes.NewReader(data))
	return err
}

func CreateADUDeployment(token string, aduAccountEndpoint string, aduAccountInstance string, group string, deployment string, aduDeployment ADUDeployment) error {
	getUrl := fmt.Sprintf("https://%s/deviceupdate/%s/management/groups/%s/deployments/%s?api-version=2021-06-01-preview", aduAccountEndpoint, aduAccountInstance, group, deployment)
	data, _ := json.Marshal(aduDeployment)
	_, err := callRESTAPI("PUT", getUrl, token, bytes.NewReader(data))
	return err
}

func DeleteADUDeployment(token string, aduAccountEndpoint string, aduAccountInstance string, group string, deployment string) error {
	getUrl := fmt.Sprintf("https://%s/deviceupdate/%s/management/groups/%s/deployments/%s?api-version=2021-06-01-preview", aduAccountEndpoint, aduAccountInstance, group, deployment)
	_, err := callRESTAPI("DELETE", getUrl, token, nil)
	return err
}
