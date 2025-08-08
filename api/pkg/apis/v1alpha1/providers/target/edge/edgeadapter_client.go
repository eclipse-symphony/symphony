// SPDX-FileCopyrightText: (C) 2024 Intel Corporation
//
// SPDX-License-Identifier: LicenseRef-Intel

package edge

import (
	"context"
	"crypto/tls"

	southbound "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/edge/api/edge_adapter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func NewEdgeAdapterClient(ctx context.Context, token string, tlsCredentials *tls.Config) (southbound.EdgeAdapterServiceClient, error) {
	addr := "192.168.200.99:6201"

	// Remove protocol prefix if present
	if len(addr) > 8 && addr[:8] == "https://" {
		addr = addr[8:]
	} else if len(addr) > 7 && addr[:7] == "http://" {
		addr = addr[7:]
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCredentials)),
	}

	// Add custom header to the gRPC request
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&metadata.MD{"content-type": []string{"application/grpc"}})))

	if token != "" {
		conn, err := grpc.DialContext(ctx, addr, opts...)
		if err != nil {
			return nil, err
		}
		return southbound.NewEdgeAdapterServiceClient(conn), nil
	} else {
		conn, err := grpc.DialContext(ctx, addr, opts...)
		if err != nil {
			return nil, err
		}
		return southbound.NewEdgeAdapterServiceClient(conn), nil
	}
}
