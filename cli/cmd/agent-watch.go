/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eclipse-symphony/symphony/cli/config"
	"github.com/eclipse-symphony/symphony/cli/utils"
	"github.com/spf13/cobra"
)

var (
	watchInterval time.Duration
	watchTimeout  time.Duration
	watchCount    int
)

var AgentWatchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch reported status properties on a registered Target",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateAgentTarget(); err != nil {
			return err
		}
		if watchInterval <= 0 || watchTimeout < 0 || watchCount < 0 {
			return fmt.Errorf("watch interval must be positive; timeout and count must not be negative")
		}
		settings, err := resolveCurrentContext()
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if watchTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, watchTimeout)
			defer cancel()
		}
		return watchAgentStatus(ctx, cmd, settings, watchInterval, watchCount)
	},
}

func watchAgentStatus(ctx context.Context, cmd *cobra.Command, settings config.MaestroContext, interval time.Duration, count int) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	last := ""
	observed := 0
	fmt.Fprintf(cmd.ErrOrStderr(), "Watching Target %s/%s; waiting for status reports...\n", targetNamespace, targetName)
	for {
		requestCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		objects, err := utils.GetInNamespace(requestCtx, settings.Url, settings.User, settings.Secret, "target", "", "", targetName, targetNamespace)
		cancel()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil
			}
			if ctx.Err() != nil {
				return fmt.Errorf("status watch ended after %d reports: %w", observed, ctx.Err())
			}
			return fmt.Errorf("read Target status (register the Target first): %w", err)
		}
		if len(objects) != 1 {
			return fmt.Errorf("expected one Target, received %d", len(objects))
		}
		data, err := json.Marshal(objects[0])
		if err != nil {
			return err
		}
		var target Target
		if err := json.Unmarshal(data, &target); err != nil {
			return err
		}
		properties := target.Status.Properties
		if properties["agent.reportedAt"] != "" {
			data, err := json.MarshalIndent(properties, "", "  ")
			if err != nil {
				return err
			}
			if string(data) != last {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(data)); err != nil {
					return err
				}
				last = string(data)
				observed++
				if count > 0 && observed >= count {
					return nil
				}
			}
		}
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil
			}
			return fmt.Errorf("status watch ended after %d reports: %w", observed, ctx.Err())
		case <-ticker.C:
		}
	}
}

func init() {
	AgentWatchCmd.Flags().DurationVar(&watchInterval, "watch-interval", time.Second, "Target status polling interval")
	AgentWatchCmd.Flags().DurationVar(&watchTimeout, "timeout", 0, "Stop with an error if this duration elapses (0 waits until Ctrl+C)")
	AgentWatchCmd.Flags().IntVar(&watchCount, "count", 0, "Exit successfully after this many distinct reports (0 watches continuously)")
	AgentCmd.AddCommand(AgentWatchCmd)
}
