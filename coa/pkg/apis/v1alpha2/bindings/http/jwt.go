/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"

	v1alpha2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/valyala/fasthttp"
	v1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type JWT struct {
	AuthHeader       string                 `json:"authHeader"`
	VerifyKey        string                 `json:"verifyKey"`
	MustHave         []string               `json:"mustHave,omitempty"`
	MustMatch        map[string]interface{} `json:"mustMatch,omitempty"`
	AuthServer       AuthServer             `json:"authServer,omitempty"`
	verifyKey        *rsa.PublicKey
	IgnorePaths      []string          `json:"ignorePaths,omitempty"`
	Roles            []ClaimRoleMap    `json:"roles,omitempty"`
	EnableRBAC       bool              `json:"enableRBAC,omitempty"`
	Policy           map[string]Policy `json:"policy,omitempty"`
	DisableUserCreds bool              `json:"disableUserCreds,omitempty"`
}

// enum string for AuthServer
type AuthServer string

const (
	// AuthServerKuberenetes means we are using kubernetes api server as auth server
	AuthServerKuberenetes AuthServer = "kubernetes"
	SymphonyIssuer        string     = "symphony"
	caFileName            string     = "/etc/symphony-api/tls/tls.crt"
)

var (
	symphonyAPIAddressBase       = os.Getenv("SYMPHONY_API_URL")
	namespace                    = os.Getenv("POD_NAMESPACE")
	apiServiceAccountName        = os.Getenv("SERVICE_ACCOUNT_NAME")
	controllerServiceAccountName = os.Getenv("SYMPHONY_CONTROLLER_SERVICE_ACCOUNT_NAME")
	subjectList                  = getSubjectList()
	ServiceName                  = os.Getenv("SYMPHONY_SERVICE_NAME")
)

func getSubjectList() []string {
	subjects := os.Getenv("CLIENT_SUBJECTS")

	// Split the subjects string by semicolon
	additionalSubjects := strings.Split(subjects, ";")
	return additionalSubjects
}

func isSubjectValid(subject string) bool {
	for _, s := range subjectList {
		if strings.Contains(subject, s) {
			// The subject contains one of the valid subjects
			return true
		}
	}
	return false
}
func loadCACertPool(caFile string) (*x509.CertPool, error) {
	caCertPool := x509.NewCertPool()
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA file %s: %v", caFile, err)
	}
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate from file %s", caFile)
	}

	return caCertPool, nil
}

func getApiServiceAccountUsername() (string, error) {
	if namespace == "" || apiServiceAccountName == "" {
		return "", v1alpha2.NewCOAError(nil, "Unable to retrieve environment variables for api service account", v1alpha2.InternalError)
	}
	return fmt.Sprintf("system:serviceaccount:%s:%s", namespace, apiServiceAccountName), nil
}

func getControllerServiceAccountUsername() (string, error) {
	if namespace == "" || controllerServiceAccountName == "" {
		return "", v1alpha2.NewCOAError(nil, "Unable to retrieve environment variables for controller service account", v1alpha2.InternalError)
	}
	return fmt.Sprintf("system:serviceaccount:%s:%s", namespace, controllerServiceAccountName), nil
}

type ClaimRoleMap struct {
	Role  string `json:"role"`
	Claim string `json:"claim"`
	Value string `json:"value"`
}
type Policy struct {
	Items map[string]string `json:"items"`
}

func (j JWT) JWT(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if j.IgnorePaths != nil {
			for _, p := range j.IgnorePaths {
				if p == string(ctx.Path()) {
					next(ctx)
					return
				}
			}
		}
		if ctx.IsOptions() {
			next(ctx)
			return
		}

		tokenStr := j.readAuthHeader(ctx)
		if tokenStr == "" {
			// Cert based auth
			conn := ctx.Conn()
			tlsConn, ok := conn.(*tls.Conn)
			if !ok {
				ctx.Error("Forbidden", fasthttp.StatusForbidden)
				return
			}

			// Get the TLS connection state
			state := tlsConn.ConnectionState()
			if len(state.PeerCertificates) == 0 {
				ctx.Error("Client certificate required", fasthttp.StatusUnauthorized)
				return
			}

			// Get the client certificate
			clientCert := state.PeerCertificates[0]
			subjectName := clientCert.Subject.String()
			log.ErrorfCtx(ctx, "JWT: Token is empty.")
			// Verify the client certificate
			caCertPool, err := loadCACertPool(caFileName)
			if err != nil {
				log.ErrorfCtx(ctx, "JWT: Could not load CA cert pool. %s\n", err.Error())
				ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
				return
			}
			opts := x509.VerifyOptions{
				Roots: caCertPool,
			}
			if _, err := clientCert.Verify(opts); err != nil {
				log.InfofCtx(ctx, "JWT: The cert is not a symphony working cert. It is a bootstrap cert.")
				subjectValid := isSubjectValid(subjectName)
				if !subjectValid {
					log.ErrorfCtx(ctx, fmt.Sprintf("JWT: The cert has no valid subject name, %s.", subjectName))
					ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
					return
				} else {
					uri := ctx.Request.URI().String()
					if strings.Contains(uri, "/targets/getcert") || strings.Contains(uri, "/files") {
						next(ctx)
					} else {
						log.ErrorfCtx(ctx, "JWT: Bootstrap cert can only access getcert, and files endpoints.")
						ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
						return
					}
				}
			} else {
				if strings.Contains(subjectName, ServiceName) {
					next(ctx)
				} else {
					log.ErrorfCtx(ctx, "JWT: The cert should have a valid symphony working cert subject.")
					ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
				}
			}
		} else {
			issuer, err := decodeJWTTokenForIssuer(tokenStr, ctx)
			if err != nil {
				log.ErrorfCtx(ctx, "JWT: Could not decode issuer from token. %s", err.Error())
				ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
				return
			}
			if issuer == SymphonyIssuer {
				if j.DisableUserCreds == true {
					log.Infof("JWT: Token with username plus pwd is not allowed.")
					ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
					return
				}
				log.Debugf("JWT: Validating token with username plus pwd.")
				_, roles, err := j.validateToken(tokenStr)
				if err != nil {
					log.ErrorCtx(ctx, "JWT: Validate token with user creds failed. %s", err.Error())
					ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
					return
				} else {
					if j.EnableRBAC {
						path := string(ctx.Path())
						method := string(ctx.Method())
						for _, role := range roles {
							if v, ok := j.Policy[role]; ok {
								for key, val := range v.Items {
									if key == "*" || strings.HasPrefix(path, key) {
										if val == "*" || strings.Contains(val, method) {
											next(ctx)
											return
										}
									}
								}
							}
						}
						ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
						return
					}
					next(ctx)
				}
			} else {
				if j.AuthServer == AuthServerKuberenetes {
					log.DebugfCtx(ctx, "JWT: Validating token with k8s.")
					err := j.validateServiceAccountToken(ctx, tokenStr)
					if err != nil {
						log.ErrorfCtx(ctx, "JWT: Validate token with k8s failed. %s", err.Error())
						ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
						return
					}
					next(ctx)
				} else {
					log.ErrorfCtx(ctx, "JWT: Not supported auth server, %s.", j.AuthServer)
					ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
					return
				}
			}
		}
	}
}
func (j JWT) readAuthHeader(ctx *fasthttp.RequestCtx) string {
	v := ctx.Request.Header.Peek(j.AuthHeader)
	if v != nil {
		tokenStr := string(v)
		token := strings.Split(tokenStr, "Bearer ")
		if len(token) == 2 {
			return strings.TrimSpace(token[1])
		} else {
			return ""
		}
	}
	return ""
}
func (j *JWT) validateToken(tokenStr string) (map[string]interface{}, []string, error) {
	ret := make(map[string]interface{})
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if j.verifyKey != nil {
				return j.verifyKey, nil
			} else {
				if strings.HasPrefix(j.VerifyKey, "-----BEGIN PUBLIC KEY-----") {
					verifyKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(j.VerifyKey))
					if err != nil {
						return ret, v1alpha2.NewCOAError(nil, "failed to parse public key", v1alpha2.BadConfig)
					}
					j.verifyKey = verifyKey
					return j.verifyKey, nil
				} else {
					return []byte(j.VerifyKey), nil
				}
			}
		},
	)
	if err != nil {
		return ret, nil, err
	}
	if !token.Valid {
		return ret, nil, errors.New("invalid token")
	}
	for k, v := range claims {
		ret[k] = v
	}
	if j.MustHave != nil && len(j.MustHave) > 0 {
		for _, k := range j.MustHave {
			if _, ok := ret[k]; !ok {
				return ret, nil, fmt.Errorf("required claim '%s' is not found", k)
			}
		}
	}
	if j.MustMatch != nil && len(j.MustMatch) > 0 {
		for k, v := range j.MustMatch {
			if hv, ok := ret[k]; ok {
				if hv != v {
					return ret, nil, fmt.Errorf("claim '%s' doesn't have required value", k)
				}
			} else {
				return ret, nil, fmt.Errorf("required claim '%s' is not found", k)
			}
		}
	}
	var roles []string
	if j.EnableRBAC {
		roles = make([]string, 0)
		for _, m := range j.Roles {
			if v, ok := ret[m.Claim]; ok {
				if m.Value == "*" || v == m.Value {
					roles = append(roles, m.Role)
				}
			}
		}

	}
	return ret, roles, nil
}

func decodeJWTTokenForIssuer(tokenString string, ctx context.Context) (string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		issuer, ok := claims["iss"].(string)
		if !ok {
			log.DebugfCtx(ctx, "The iss claim is not a string")
			return "", errors.New("the iss claim is not a string")
		}
		log.DebugfCtx(ctx, "Issuer: %s", issuer)
		return issuer, nil
	} else {
		log.DebugfCtx(ctx, "Invalid token")
		return "", errors.New("invalid token")
	}
}

func (j *JWT) validateServiceAccountToken(ctx *fasthttp.RequestCtx, tokenStr string) error {
	clientset, err := getKubernetesClient()
	if err != nil {
		log.ErrorfCtx(ctx, "JWT: Could not initialize Kubernetes client.\n")
		return v1alpha2.NewCOAError(err, "Could not initialize Kubernetes client", v1alpha2.InternalError)
	}
	tokenReview := &v1.TokenReview{
		Spec: v1.TokenReviewSpec{
			Token: tokenStr,
			Audiences: []string{
				symphonyAPIAddressBase,
			},
		},
	}

	result, err := clientset.AuthenticationV1().TokenReviews().Create(ctx, tokenReview, metav1.CreateOptions{})
	if err != nil {
		log.ErrorfCtx(ctx, "JWT: Token review using kubernetes api server failed. %s\n", err.Error())
		return v1alpha2.NewCOAError(err, "Token review using kubernetes api server failed.", v1alpha2.InternalError)
	}
	if !result.Status.Authenticated {
		log.ErrorfCtx(ctx, "JWT: Validate token with k8s failed. K8s returned not authenticated.\n")
		return v1alpha2.NewCOAError(nil, "Authentication failed.", v1alpha2.Unauthorized)
	} else {
		apiUsername, err := getApiServiceAccountUsername()
		if err != nil {
			return err
		}
		controllerUsername, err := getControllerServiceAccountUsername()
		if err != nil {
			return err
		}
		if result.Status.User.Username != apiUsername && result.Status.User.Username != controllerUsername {
			log.ErrorfCtx(ctx, "JWT: Validate token with k8s failed. K8s returned invalid username, %s\n", result.Status.User.Username)
			return v1alpha2.NewCOAError(nil, "Authentication failed.", v1alpha2.Unauthorized)
		}
	}
	return nil

}
func getKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
