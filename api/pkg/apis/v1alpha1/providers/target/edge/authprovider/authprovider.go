package authprovider

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TokenResponseModel struct {
	AccessToken         string `json:"accessToken"`
	AccessTokenLifeTime int64  `json:"accessTokenLifeTime"`
	TokenType           string `json:"tokenType"`
}

type CaeAuthenticateRequest struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
}

type AuthenticationService struct {
	client      *http.Client
	Credentials *tls.Config
}

func ParseCertificatePEM(certPEM string) ([][]byte, error) {
	var certs [][]byte
	block, rest := pem.Decode([]byte(certPEM))
	for block != nil {
		certs = append(certs, block.Bytes)
		block, rest = pem.Decode(rest)
	}
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}
	return certs, nil
}

// NewAuthenticationService creates a new instance of AuthenticationService with TLS credentials.
func NewAuthenticationService() (*AuthenticationService, error) {
	certBytes, err := ParseCertificatePEM(Certificate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate PEM: %v", err)
	}

	cert := tls.Certificate{
		Certificate: certBytes,
	}

	credentials := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            nil,
		InsecureSkipVerify: true,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: credentials,
		},
	}

	return &AuthenticationService{client: client, Credentials: credentials}, nil
}

// GetSessionIdAsync authenticates the user and retrieves a session ID asynchronously.
func (s *AuthenticationService) GetSessionIdAsync(baseURL string) (string, time.Time, error) {
	authRequest := CaeAuthenticateRequest{
		Username: Username,
		Password: Password,
	}

	jsonContent, err := json.Marshal(authRequest)
	if err != nil {
		return "", time.Time{}, err
	}
	posturl := fmt.Sprintf("%s/login", baseURL)
	// log.Printf("Sending the login request to URL : %s", posturl)
	req, err := http.NewRequest("POST", posturl, bytes.NewBuffer(jsonContent))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error creating new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// log.Printf("Response status: %s", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("authentication failed: %s", resp.Status)
	}

	var sessionId string
	var expires string
	cookies := resp.Header["Set-Cookie"]
	for _, cookie := range cookies {
		parts := strings.Split(cookie, ";")
		for _, part := range parts {
			if strings.HasPrefix(part, "sessionId=") {
				sessionIdUrlEncoded := strings.TrimPrefix(part, "sessionId=")
				sessionId, err = url.QueryUnescape(sessionIdUrlEncoded)
				if err != nil {
					sessionId = sessionIdUrlEncoded
				}
			} else if strings.HasPrefix(part, " expires=") {
				expires = strings.TrimPrefix(part, " expires=")
			}
		}
	}

	layout := "Mon, 02 Jan 2006 15:04:05 MST"
	parsedTime, _ := time.Parse(layout, expires)
	return sessionId, parsedTime, nil
}

// AuthenticateAsync authenticates the user and retrieves an access token asynchronously.
func (s *AuthenticationService) AuthenticateAsync(username, password string) (string, error) {
	authRequest := CaeAuthenticateRequest{
		Username: base64.StdEncoding.EncodeToString([]byte(username)),
		Password: base64.StdEncoding.EncodeToString([]byte(password)),
	}

	jsonContent, err := json.Marshal(authRequest)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://localhost:6201/identityapp:services:IdentityService/v1/authenticate/token", bytes.NewBuffer(jsonContent))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication failed: %s", resp.Status)
	}

	var tokenResponse TokenResponseModel
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

// SendTokenAsync sends the token for introspection asynchronously.
func (s *AuthenticationService) SendTokenAsync(token string) (string, error) {
	tokenRequest := map[string]string{"Token": token}

	jsonContent, err := json.Marshal(tokenRequest)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://localhost:6201/identityapp:services:IdentityService/v1/token/introspect", bytes.NewBuffer(jsonContent))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token introspection failed: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

var (
	Username = "service"
	Password = "Test@123"

	Certificate string = `-----BEGIN CERTIFICATE-----
MIIFNzCCAx+gAwIBAgIUB3lPYTe7o0hbdT4cg5/RjVcJWwwwDQYJKoZIhvcNAQEL
BQAwODElMCMGA1UECgwcU0NITkVJREVSIEVMRUNUUklDIFVTQSwgSU5DLjEPMA0G
A1UEAwwGRUFFUDI1MB4XDTI1MDUyODExMTkyN1oXDTI2MDUyODExMTkyN1owODEl
MCMGA1UECgwcU0NITkVJREVSIEVMRUNUUklDIFVTQSwgSU5DLjEPMA0GA1UEAwwG
RUFFUDI1MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAwFonKHpM++lY
pp/I/JNo3pWnYDxoWwSnZJktxxbzoweEQFxDt2R0/RDaGRDRYJjFHiyppLpav6ZO
Xv2X7rHaqIrgR7lwQktplP7Xhun5IiEp7WeqeZm1TbYZsOcFeCXFK0GfWH+28C0R
zAPt6QMOCvJ4jybN9ziVbTVLkg889rbdJFtfqBIXzQ13ZP+cPeS5DOBG9VOviYDV
77cbScGHu1DquL21ACNTof8xLCK30myvrrdH/PdRUAXhsUN/uZ8AXRDN4roerCF4
IyNlk1tJ1xTG0oGJpJGSDq3O1KZ68iPWfmZMhVQMNfV10C/CmeNdeDIXAL8Nz/6Z
ITKGMYwfHAVvq6rHc5RZxywGExuNBNtrKNz8wQrpVrBIrupUW2uMKxdklWTpGYjk
7C0mGdigKj3cAq2o+eKZ0plIgc0Sws97NNIrMyg3ymSsWUaBqM72KjQ06bgv56oe
vT9XBTosZKmMXq9XkwKVrxvO9GUpM9fa16z61tYqdKeuQNq8jRsN/nAy/Nyv/zyz
8XlxPW8IV6t5b0jIMk3GPI2dlsxP7coxHS7bs2i9o2FpCEZhwjoQWjeH0NPPcI11
gADYxuJzp7KwwykJHOZ7cCsP89X7HcQhHoN6Vufm5C+i55pM6yC61nOg2RL1kyIb
YiAs9+PXtjjvKR76+Aqk5jdn1zesjqsCAwEAAaM5MDcwFwYDVR0RBBAwDocECuTq
HYIGRUFFUDI1MAwGA1UdEwEB/wQCMAAwDgYDVR0PAQH/BAQDAgXgMA0GCSqGSIb3
DQEBCwUAA4ICAQAwQI6dpU67Dzml9WMedullzhfWGu8iq768GMswHJPMAvhLE7Ux
HINq86UDWfDfm+nNlDg1eJp3okANJsBsJcD4bstP89HGTmsoU777GGQYDYZ093Xw
iqG7I0X2z6VxChw57yLOr5D/QhKMJqNJ5p6DiBqqJKWPlzKIdRiPQJD26qHyh9JA
RHLReomf28p5tHXi+nZKCUB13THdx2ZAO45qvtHw6XczCvYO1gzWc9EOj8CK2jhL
U7Dy2gvg8A4sYcmrvNpDK/edM+jrAkBrrlNgQXn13qDCkhZcol1MRXecxIjJGFTb
hF4ifF3E0H+/1KxuKXLjVIKb3Vd/fp9/i96jQKEjLdx/3158OGtFs/1ug6FF9KaJ
E1YtFfO6diUwDeaERCgrob27uh6UDiBTxWOSlHNauCB68X3uSzK8OtBb9zhinQN/
Q47og0kN76btixMPYwa2fWSiX4eGewbOLTWsHYQLhNjt+QK1+AKqQcvZkuvX644q
gQ5SAYgKaQkJpHGjNUMll7dFN3HsLN7tNXisNY2+eGrq5uqbPZpJ52XoL381ScaA
Q8iWEZ9xVeV9Mx9a7Ku6gCX7/b391gS0fPu6h8wXnaRg0IJ+60t2i5IPMhfhAyLR
xfv0kYHav18KlwBkM/Y6Z8KoZZmjs/xi+h/tDoCrgIzovtm1AB/8UWTN2A==
-----END CERTIFICATE-----`
)
