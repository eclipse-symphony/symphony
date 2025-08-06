// SPDX-FileCopyrightText: (C) 2024 Intel Corporation
//
// SPDX-License-Identifier: LicenseRef-Intel

package edge

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/edge/api/system_model"
	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type tokenAuth struct {
	token                    string
	requireTransportSecurity bool
}

func (t tokenAuth) RequireTransportSecurity() bool {
	return t.requireTransportSecurity
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + t.token}, nil
}

func NewSystemModelClient(ctx context.Context, token string, tlsCredentials *tls.Config) (system_model.SystemModelClient, error) {
	addr, ok := os.LookupEnv("MODEL_SERVICE_ADDRESS")
	if !ok {
		return nil, fmt.Errorf("system model service address is not set")
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCredentials)),
	}
	// Add custom header to the gRPC request
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&metadata.MD{"content-type": []string{"application/grpc"}})))
	if token != "" {
		conn, err := grpc.DialContext(ctx, addr, opts...,
		//grpc.WithTransportCredentials(credentials.NewTLS(tlsCredentials)),
		// grpc.WithStreamInterceptor(retry.RetryingStreamClientInterceptor(retry.WithRetryOn(codes.Unavailable, codes.Unknown))),
		// grpc.WithUnaryInterceptor(retry.RetryingUnaryClientInterceptor(retry.WithRetryOn(codes.Unavailable, codes.Unknown))),
		// grpc.WithPerRPCCredentials(tokenAuth{token: token, requireTransportSecurity: true}),
		)
		if err != nil {
			return nil, err
		}
		return system_model.NewSystemModelClient(conn), nil
	} else {
		conn, err := grpc.DialContext(ctx, addr, opts...,
		// grpc.WithTransportCredentials(insecure.NewCredentials()),
		// grpc.WithStreamInterceptor(retry.RetryingStreamClientInterceptor(retry.WithRetryOn(codes.Unavailable, codes.Unknown))),
		// grpc.WithUnaryInterceptor(retry.RetryingUnaryClientInterceptor(retry.WithRetryOn(codes.Unavailable, codes.Unknown))),
		// grpc.WithPerRPCCredentials(tokenAuth{token: "", requireTransportSecurity: false}),
		)
		if err != nil {
			return nil, err
		}
		return system_model.NewSystemModelClient(conn), nil
	}

}
