package helm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	cerrors "github.com/containerd/containerd/remotes/errors"
	"helm.sh/helm/v3/pkg/registry"
)

type (
	tokenExchangeResponse struct {
		RefreshToken string `json:"refresh_token"`
	}

	tokenExchangeRequest struct {
		GrantType   string `json:"grant_type"`
		Service     string `json:"service"`
		AccessToken string `json:"access_token"`
	}
)

const (
	defaultAuthUser   = "00000000-0000-0000-0000-000000000000"
	defaultAzureScope = "https://management.azure.com/.default"
	azureCrPostfix    = ".azurecr.io"
	exchangeURLFormat = "https://%s/oauth2/exchange"
)

// loginToACR logs in to an Azure Container Registry using the provided helm registry client.
func loginToACR(ctx context.Context, host string) error {
	client, err := registry.NewClient()
	if err != nil {
		sLog.ErrorfCtx(ctx, "Failed to create registry client: %+v", err)
		return err
	}

	cred, err := azidentity.NewManagedIdentityCredential(nil)
	if err != nil {
		sLog.ErrorfCtx(ctx, "failed to obtain a credential: %v", err)
		return err
	}
	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{defaultAzureScope},
	})

	if err != nil {
		sLog.ErrorfCtx(ctx, "failed to get token: %v", err)
		return err
	}

	acrToken, err := exchangeToken(host, token.Token)
	if err != nil {
		sLog.ErrorfCtx(ctx, "failed to exchange token: %v", err)
		return err
	}

	return client.Login(host, registry.LoginOptBasicAuth(defaultAuthUser, acrToken))
}

// isUnauthorized returns true if the error is an unauthorized error from the helm sdk.
func isUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	var unexpectedStatusError = &cerrors.ErrUnexpectedStatus{}
	if errors.As(err, unexpectedStatusError) {
		return unexpectedStatusError.StatusCode == http.StatusUnauthorized
	}
	return false
}

// isAzureContainerRegistry returns true if the host is an Azure Container Registry.
// This is a very rudimentary check that only checks for the .azurecr.io suffix.
func isAzureContainerRegistry(host string) bool {
	return strings.HasSuffix(host, azureCrPostfix)
}

func getHostFromOCIRef(ref string) (string, error) {
	if !strings.HasPrefix(ref, "oci://") {
		ref = fmt.Sprintf("oci://%s", ref)
	}
	parsed, err := url.Parse(ref)
	if err != nil {
		return "", err
	}

	return parsed.Host, nil
}

// exchangeToken exchanges an Azure AD token for an ACR refresh token.
// This is used by the Helm registry client to authenticate to ACR.
func exchangeToken(host, token string) (string, error) {
	req := tokenExchangeRequest{
		GrantType:   "access_token",
		Service:     host,
		AccessToken: token,
	}

	res := tokenExchangeResponse{}

	jsonResponse, err := http.PostForm(fmt.Sprintf(exchangeURLFormat, host), req.ToFormValues())
	if err != nil {
		return "", err
	}

	if err := json.NewDecoder(jsonResponse.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.RefreshToken, nil
}

func (r *tokenExchangeRequest) ToFormValues() url.Values {
	return url.Values{
		"grant_type":   {r.GrantType},
		"service":      {r.Service},
		"access_token": {r.AccessToken},
	}
}
