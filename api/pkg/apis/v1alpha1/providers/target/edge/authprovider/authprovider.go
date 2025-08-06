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

	Certificate string =
	`-----BEGIN CERTIFICATE-----
	MIIFUjCCAzqgAwIBAgIUCwB/DBO6/7e92mjR7RFEYu/O9Q8wDQYJKoZIhvcNAQEL
	BQAwQTElMCMGA1UECgwcU0NITkVJREVSIEVMRUNUUklDIFVTQSwgSU5DLjEYMBYG
	A1UEAwwPREVTS1RPUC1VRDQ4NEMyMB4XDTI1MDIxMDEwMzI1NVoXDTI2MDIxMDEw
	MzI1NVowQTElMCMGA1UECgwcU0NITkVJREVSIEVMRUNUUklDIFVTQSwgSU5DLjEY
	MBYGA1UEAwwPREVTS1RPUC1VRDQ4NEMyMIICIjANBgkqhkiG9w0BAQEFAAOCAg8A
	MIICCgKCAgEAkyzJbPkO8hN7Xk6+KlG507oP0yl9K8vQSF3tvoWiwluhkr7zcPhb
	DNcBgMwq0g7pFVC8Oqkj6JKgc6DyFGfSzEbhxtw3kifKGUPcB1Yzsd7XuBXqwHZO
	KQKCFJk5CetnKxRVI1sZSBud7VVzmrWNBzSRI0KX+VOIUwo8SRvf9WbX85QTOVaB
	z+/TTEmLknLx5KfLecsZ43pELKwCN50dhDvMJQEcSQNAue3fV1DjfTLNN86LS0e7
	mXY2BEterRKN6uvMsE1ytlIE2r/rvnFnesfUp0+3cCdz0tL8x+E2xeCBpmfEnjT5
	RDnhDDL7RT8C9Hdq4fi2JI5c4nDvchTyL61JPirc9sSJlkTY7h09vny+zXP8PPAk
	9uHNhTccSFx5sFnUGxmW6spmBsnNv5Nd753xFG7J9a8lcovC1lZA267c8Vl7QWIL
	dQGjNaYytcHoHjq6NlFTrRFCwsOxaqW5a8io+s/GOpSBjR+nqSlZ8MZagOUQu8Sp
	0RERaFyj/PCmXe67/AwIEJIcB0J8dFnIGQ4h40GyQ9nT9P+loRWDrkNHxPt6xtJ/
	CW1naPDPqPW0y3h+ysSnFLDgkpv7T/s1miuangnzl4Hj4gira3MsmBPaq8HL+rtN
	30l4WBFwGbgRs+ETidzU1AqR3bAUTHAKNWqPJN1SoTMCWhxLBJZ3CM8CAwEAAaNC
	MEAwIAYDVR0RBBkwF4cECuTqFIIPREVTS1RPUC1VRDQ4NEMyMAwGA1UdEwEB/wQC
	MAAwDgYDVR0PAQH/BAQDAgXgMA0GCSqGSIb3DQEBCwUAA4ICAQBiMvaMmKt7Havx
	GJlYgYV55mHjgTIPOC/IRozpy+02djaubolP/Sh5KzxM9yjBDUhXyFNw+pPl4Flq
	Ef8fW9YWhVTQzMn5SdSAedDNLhtCGT+f8iVZwKHs+dOs5jrfjMM/LXYCOt6o1yr1
	mkqFCte3cId+LirSdU4TEUIJAc5Z1ek5flbk7P+1qQeBUrp0DtEHaBdkHzJR8P+F
	QKgELXhfBgF96m0hNoGV4TFSYM0s5pVIrVbDbAB3ImdYQjfahGoQWRTr7PxrVCBK
	BOmkSDf9bFuB5P87o0hP5qiJM/psQeRgDT2gT3QUlgo3kx70Ep0cRxmhw3TPQqXl
	WNBQYFAebTegnjTV39gY+1gAz7UC0qWsWIE4jUQQzTy44xvpDEv19ENPUvUHwBDS
	VSuA19mmyEmB/Up3AhK9mN6n+9i80kAc0gQ5BM1+qOJADQEkNboISqGO18CHEAQ1
	QORjJR2hcOnaP5qqSyzrSTgH0vc9y7BjQbxDF0J/QyzSLgybZY/3wyr4D7Qqgeed
	VCG852YmXWLDQukEVphpBjsX2xYJ74bFl+6+L8ujQz+WFY4XvEebSOOL9qA3vuon
	MJzF0PuHVoD6jkA5tBVymiX/eAexMsojYppSXGPfKmdIIQejgVuhRywZMBRHThTS
	VoWSq2KqQpZo9URVn4W9ookHRda9zA==
	-----END CERTIFICATE-----'`
)

