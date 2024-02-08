/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func generateJWTToken(signingKey interface{}, method jwt.SigningMethod, userName string, expiresAt time.Time, issuedAt time.Time, notAfter time.Time, issuer string, subject string, audiences []string) (string, error) {
	claims := TestCustomClaims{
		User: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(notAfter),
			Issuer:    issuer,
			Subject:   subject,
			Audience:  jwt.ClaimStrings(audiences),
		},
	}
	token := jwt.NewWithClaims(method, claims)
	return token.SignedString(signingKey)
}

func TestValidateWithValidTokenAuthN(t *testing.T) {
	j := JWT{
		AuthHeader: "Authorization",
		VerifyKey:  "test",
		EnableRBAC: false,
	}

	token, err := generateJWTToken([]byte("test"), jwt.SigningMethodHS256, "test", time.Now().Add(time.Hour), time.Now(), time.Now(), "test", "test", []string{"test"})
	assert.Nil(t, err)
	_, _, err = j.validateToken(token)
	assert.Nil(t, err)
}

func TestValidateWithInvalidTokenAuthN(t *testing.T) {
	j := JWT{
		AuthHeader: "Authorization",
		VerifyKey:  "test",
		EnableRBAC: false,
	}

	// expiresAt is in the past, notAfter is in the past
	// token is expired
	token, err := generateJWTToken([]byte("test"), jwt.SigningMethodHS256, "test", time.Now().Add(-1*time.Hour), time.Now(), time.Now().Add(-1*time.Hour), "test", "test", []string{"test"})
	assert.Nil(t, err)

	_, _, err = j.validateToken(token)
	assert.NotNil(t, err)
}

func TestValidateWithTokenMustHaveNotExists(t *testing.T) {
	j := JWT{
		AuthHeader: "Authorization",
		VerifyKey:  "test",
		EnableRBAC: false,
		MustHave:   []string{"username"},
	}

	// TestClaims has user claim, but MustHave is username
	token, err := generateJWTToken([]byte("test"), jwt.SigningMethodHS256, "test", time.Now().Add(time.Hour), time.Now(), time.Now(), "test", "test", []string{"test"})
	assert.Nil(t, err)

	_, _, err = j.validateToken(token)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "required claim 'username' is not found")
}

func TestValidateWithTokenMustMatchNotExists(t *testing.T) {
	j := JWT{
		AuthHeader: "Authorization",
		VerifyKey:  "test",
		EnableRBAC: false,
		MustMatch:  map[string]interface{}{"username": "test"},
	}

	// TestClaims has user claim, but MustMatch is username
	token, err := generateJWTToken([]byte("test"), jwt.SigningMethodHS256, "test", time.Now().Add(time.Hour), time.Now(), time.Now(), "test", "test", []string{"test"})
	assert.Nil(t, err)

	_, _, err = j.validateToken(token)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "required claim 'username' is not found")
}

func TestValidateWithTokenMustMatchNotMatch(t *testing.T) {
	j := JWT{
		AuthHeader: "Authorization",
		VerifyKey:  "test",
		EnableRBAC: false,
		MustMatch:  map[string]interface{}{"user": "test1"},
	}

	// TestClaims has user claim, but MustMatch is user with different value
	token, err := generateJWTToken([]byte("test"), jwt.SigningMethodHS256, "test", time.Now().Add(time.Hour), time.Now(), time.Now(), "test", "test", []string{"test"})
	assert.Nil(t, err)

	_, _, err = j.validateToken(token)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "claim 'user' doesn't have required value")
}

func TestValidateWithTokenWithCert(t *testing.T) {
	// generate a new private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.Nil(t, err)

	// Marshal the public key into PKIX, ASN.1 DER form.
	pubASN1, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}

	// PEM encode the public key.
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})

	j := JWT{
		AuthHeader: "Authorization",
		VerifyKey:  string(pubBytes),
		EnableRBAC: false,
	}

	token, err := generateJWTToken(privateKey, jwt.SigningMethodRS256, "test", time.Now().Add(time.Hour), time.Now(), time.Now(), "test", "test", []string{"test"})
	assert.Nil(t, err)
	_, _, err = j.validateToken(token)
	assert.Nil(t, err)
}
