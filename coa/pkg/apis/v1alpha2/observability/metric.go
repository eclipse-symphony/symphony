/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package observability

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

type (
	// Metrics provides the implementation for dealing with metrics. Metrics are
	// used to measure quantitative data over time, such as the rate of requests
	// or the memory usage of a system.
	Metrics interface {
		// Counter returns a new Counter.
		Counter(
			name, description string,
			attrs ...map[string]any,
		) (Counter, error)

		// Gauge returns a new Gauge. Attributes applied at the gauge level will
		// be applied to every value emitted; attributes applied at the value
		// level will take precedence. Each unique value combination of
		// attributes will be emitted as separate values so if you want values
		// to be grouped together, they must have the same attribute set.
		Gauge(
			name, description string,
			attrs ...map[string]any,
		) (Gauge, error)

		// Histogram returns a new Histogram.
		Histogram(
			name, description string,
			attrs ...map[string]any,
		) (Histogram, error)
	}
	metrics struct {
		provider otelmetric.MeterProvider
		meter    otelmetric.Meter
	}

	// Counter is an instrument to be used to record the sum of float64
	// values once per measurement cycle and then resetting. No trace context.
	Counter interface {
		// Add adds the current value of the counter.
		Add(incr float64, attrs ...map[string]any)

		// Close stops the counter from emitting values after the next
		// measurement cycle.
		Close()
	}
	counter struct {
		reg        otelmetric.Registration
		values     map[attribute.Set]float64
		attrs      map[string]any
		mu         sync.Mutex
		closed     bool
		closedEmit chan struct{}
	}

	// Gauge is a instrument to be used to emit values once per measurement
	// cycle and then resetting.
	Gauge interface {
		// Set sets the current value of the gauge.
		Set(value float64, attrs ...map[string]any)

		// Close stops the gauge from emitting values after the next measurement
		// cycle.
		Close()
	}
	gauge struct {
		reg        otelmetric.Registration
		values     map[attribute.Set]float64
		attrs      map[string]any
		mu         sync.Mutex
		closed     bool
		closedEmit chan struct{}
	}

	// Histogram is a instrument to be used to record the distribution of
	// float64 measurements.
	Histogram interface {
		// Add records a change to the instrument.
		Add(incr int64, attrs ...map[string]any)
	}
	histogram struct {
		attrs map[string]any
		h     otelmetric.Int64Histogram
	}
)

func (m *metrics) Counter(
	name, description string,
	attrs ...map[string]any,
) (Counter, error) {
	i, err := m.meter.Float64ObservableUpDownCounter(
		name,
		otelmetric.WithDescription(description),
	)
	if err != nil {
		return nil, errors.New("Counter cannot be created")
	}

	c := &counter{
		attrs:      mergeAttrs(attrs...),
		values:     make(map[attribute.Set]float64),
		closedEmit: make(chan struct{}, 1),
	}
	r, err := m.meter.RegisterCallback(
		func(ctx context.Context, o otelmetric.Observer) error {
			c.mu.Lock()
			defer c.mu.Unlock()

			for k, v := range c.values {
				o.ObserveFloat64(i, v, otelmetric.WithAttributes(k.ToSlice()...))
			}

			if c.closed {
				c.closedEmit <- struct{}{}
			}

			c.values = make(map[attribute.Set]float64, 0)

			return nil
		},
		i,
	)
	if err != nil {
		return nil, errors.New("Counter cannot be created")
	}

	c.reg = r

	return c, nil
}

func (c *counter) Add(
	incr float64,
	attrs ...map[string]any,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.attrs != nil {
		attrs = append([]map[string]any{c.attrs}, attrs...)
	}

	k := attribute.NewSet(convertMapToAttributes(mergeAttrs(attrs...))...)
	c.values[k] += incr
}

func (c *counter) Close() {
	// Unregister will hang if done within the callback so to ensure
	// all values are emited we listen to the channel from the called
	go func() {
		<-c.closedEmit
		c.clean()
		if err := c.reg.Unregister(); err != nil {
			fmt.Println("Error closing async counter due to: ", err)
		}
	}()

	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
}

func (c *counter) clean() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.values {
		delete(c.values, key)
	}

	for key := range c.attrs {
		delete(c.attrs, key)
	}
}

func (m *metrics) Gauge(
	name, description string,
	attrs ...map[string]any,
) (Gauge, error) {
	i, err := m.meter.Float64ObservableGauge(
		name,
		otelmetric.WithDescription(description),
	)
	if err != nil {
		return nil, errors.New("Gauge cannot be created")
	}

	g := &gauge{
		attrs:      mergeAttrs(attrs...),
		values:     make(map[attribute.Set]float64),
		closedEmit: make(chan struct{}, 1),
	}
	r, err := m.meter.RegisterCallback(
		func(ctx context.Context, o otelmetric.Observer) error {
			g.mu.Lock()
			defer g.mu.Unlock()

			for k, v := range g.values {
				o.ObserveFloat64(i, v, otelmetric.WithAttributes(k.ToSlice()...))
			}

			if g.closed {
				g.closedEmit <- struct{}{}
			}

			g.values = make(map[attribute.Set]float64, 0)

			return nil
		},
		i,
	)
	if err != nil {
		return nil, errors.New("Gauge cannot be created")
	}

	g.reg = r

	return g, nil
}

func (g *gauge) Set(
	value float64,
	attrs ...map[string]any,
) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.attrs != nil {
		attrs = append([]map[string]any{g.attrs}, attrs...)
	}

	k := attribute.NewSet(convertMapToAttributes(mergeAttrs(attrs...))...)
	g.values[k] = value
}

func (g *gauge) clean() {
	g.mu.Lock()
	defer g.mu.Unlock()

	for key := range g.values {
		delete(g.values, key)
	}

	for key := range g.attrs {
		delete(g.attrs, key)
	}
}

func (g *gauge) Close() {
	// Unregister will hang if done within the callback so to ensure
	// all values are emited we listen to the channel from the called
	go func() {
		<-g.closedEmit
		g.clean()
		if err := g.reg.Unregister(); err != nil {
			fmt.Println("Error closing gauge due to: ", err)
		}
	}()

	g.mu.Lock()
	g.closed = true
	g.mu.Unlock()
}

func (m *metrics) Histogram(
	name, description string,
	attrs ...map[string]any,
) (Histogram, error) {
	h, err := m.meter.Int64Histogram(
		name,
		otelmetric.WithDescription(description),
	)
	if err != nil {
		return nil, errors.New("Histogram cannot be created")
	}

	hist := &histogram{
		h:     h,
		attrs: mergeAttrs(attrs...),
	}
	return hist, nil
}

func (h *histogram) Add(
	incr int64,
	attrs ...map[string]any,
) {
	if h.attrs != nil {
		attrs = append([]map[string]any{h.attrs}, attrs...)
	}

	h.h.Record(
		context.Background(),
		incr,
		otelmetric.WithAttributes(convertMapToAttributes(mergeAttrs(attrs...))...),
	)
}

func mergeAttrs(attrs ...map[string]any) map[string]any {
	if len(attrs) == 0 {
		return nil
	}

	attr := make(map[string]any, 0)
	for _, a := range attrs {
		for k, v := range a {
			attr[k] = v
		}
	}
	return attr
}
