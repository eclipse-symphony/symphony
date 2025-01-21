/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"context"
	"fmt"

	"github.com/eclipse-symphony/symphony/cli/utils"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"
)

var (
	providerName string
	providerType string
	openAIKey    string
)

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Symphony provider implementation",
	Run: func(cmd *cobra.Command, args []string) {
		switch providerType {
		case "stage":
			generateStage(providerName, openAIKey)
		case "target":
			generateTarget(providerName, openAIKey)
		default:
			fmt.Printf("\n%s  Unrecognized proivder type (supported types: 'stage' or 'target'): %s%s\n\n", utils.ColorRed(), providerType, utils.ColorReset())
			return
		}
	},
}

func openAIComplete() (string, error) {
	client := openai.NewClient(
		option.WithAPIKey(openAIKey),
	)
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hello"),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		return "", err
	}
	return chatCompletion.Choices[0].Message.Content, nil
}

func generateStage(name, key string) {
	// Implementation
}

func generateTarget(name, key string) {
	// Implementation
}

func init() {
	GenerateCmd.Flags().StringVarP(&providerName, "name", "n", "", "Provider name")
	GenerateCmd.Flags().StringVarP(&providerType, "type", "t", "", "Provider type (stage or target)")
	GenerateCmd.Flags().StringVarP(&openAIKey, "openai-key", "k", "", "OpenAI API key")

	GenerateCmd.MarkFlagRequired("name")
	GenerateCmd.MarkFlagRequired("type")
	GenerateCmd.MarkFlagRequired("openai-key")
	RootCmd.AddCommand(GenerateCmd)
}
