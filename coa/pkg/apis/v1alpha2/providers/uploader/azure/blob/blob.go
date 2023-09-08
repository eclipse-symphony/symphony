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

package blob

import (
	"context"
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
)

type AzureBlobUploaderConfig struct {
	Name      string `json:"name"`
	Account   string `json:"account"`
	Container string `json:"container"`
}

func AzureBlobUploaderConfigFromMap(properties map[string]string) (AzureBlobUploaderConfig, error) {
	ret := AzureBlobUploaderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	if v, ok := properties["account"]; ok {
		ret.Account = utils.ParseProperty(v)
	}
	if v, ok := properties["container"]; ok {
		ret.Container = utils.ParseProperty(v)
	}
	return ret, nil
}

type AzureBlobUploader struct {
	Config  AzureBlobUploaderConfig
	Context *contexts.ManagerContext
}

func (i *AzureBlobUploader) InitWithMap(properties map[string]string) error {
	config, err := AzureBlobUploaderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (m *AzureBlobUploader) ID() string {
	return m.Config.Name
}

func (a *AzureBlobUploader) SetContext(context *contexts.ManagerContext) error {
	a.Context = context
	return nil
}

func (m *AzureBlobUploader) Init(config providers.IProviderConfig) error {
	var err error
	aConfig, err := toAzureBlobUploaderConfig(config)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "provided config is not a valid Azure blob uploader config", v1alpha2.BadConfig)
	}
	m.Config = aConfig
	return nil
}

func toAzureBlobUploaderConfig(config providers.IProviderConfig) (AzureBlobUploaderConfig, error) {
	ret := AzureBlobUploaderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	ret.Name = utils.ParseProperty(ret.Name)
	ret.Account = utils.ParseProperty(ret.Account)
	ret.Container = utils.ParseProperty(ret.Container)
	return ret, err
}

func (m *AzureBlobUploader) Upload(name string, data []byte) (string, error) {
	ctx := context.Background()
	url := "https://" + m.Config.Account + ".blob.core.windows.net/"
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return "", err
	}
	blobClient, err := blockblob.NewClient(url+m.Config.Container+"/"+name, credential, nil)
	if err != nil {
		return "", err
	}
	mime := "image/jpeg"
	_, err = blobClient.UploadBuffer(ctx, data, &blockblob.UploadBufferOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &mime,
		},
	})
	if err != nil {
		return "", err
	}
	return url + m.Config.Container + "/" + name, nil
}
