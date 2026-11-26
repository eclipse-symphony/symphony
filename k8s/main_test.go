/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestControllerCacheSyncWithoutInitialWatchBookmark(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("watch") == "true" {
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			<-r.Context().Done()
			return
		}
		if err := json.NewEncoder(w).Encode(&corev1.NamespaceList{
			TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "NamespaceList"},
			ListMeta: metav1.ListMeta{ResourceVersion: "1"},
			Items: []corev1.Namespace{{
				ObjectMeta: metav1.ObjectMeta{Name: "default", ResourceVersion: "1"},
			}},
		}); err != nil {
			t.Errorf("encode namespace list: %v", err)
		}
	}))
	defer server.Close()
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{corev1.SchemeGroupVersion})
	mapper.Add(corev1.SchemeGroupVersion.WithKind("Namespace"), meta.RESTScopeRoot)
	controllerCache, err := cache.New(&rest.Config{Host: server.URL}, cache.Options{
		Scheme: scheme, Mapper: mapper,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := controllerCache.GetInformer(ctx, &corev1.Namespace{}); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- controllerCache.Start(ctx) }()
	defer func() {
		cancel()
		if err := <-done; err != nil {
			t.Errorf("stop controller cache: %v", err)
		}
	}()
	if !controllerCache.WaitForCacheSync(ctx) {
		t.Fatal("controller cache failed to synchronize without an initial watch bookmark")
	}
	var namespace corev1.Namespace
	if err := controllerCache.Get(ctx, client.ObjectKey{Name: "default"}, &namespace); err != nil {
		t.Fatalf("webhook-style cached lookup failed: %v", err)
	}
}
