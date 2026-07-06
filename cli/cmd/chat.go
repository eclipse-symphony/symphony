/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/eclipse-symphony/symphony/cli/config"
	"github.com/eclipse-symphony/symphony/cli/utils"
	"github.com/spf13/cobra"
)

var (
	chatModel    string
	chatEndpoint string
)

// maxToolSteps caps how many tool invocations the assistant may chain before
// it must produce a final answer, preventing runaway tool loops.
const maxToolSteps = 5

const baseSystemPrompt = `You are "Symphony", the assistant for the Eclipse Symphony project - an open-source ` +
	`orchestration platform for managing applications and workloads across edge and cloud environments. ` +
	`Help the user understand and work with Symphony's concepts (such as targets, solutions, instances, ` +
	`catalogs, campaigns, activations, stages, providers, vendors and managers), its REST APIs, and the ` +
	`maestro CLI. Answer questions accurately and concisely. If a question is not related to Symphony, ` +
	`politely let the user know and steer the conversation back to Symphony topics.`

const toolProtocolPrompt = `

You can inspect and manage live Symphony objects by calling tools exposed by the Symphony MCP server. ` +
	`When you need to use a tool, do not guess or rely on memory - reply with a SINGLE line containing exactly:
#TOOL {"name":"<tool-name>","arguments":{ ... }}
and nothing else (no code fences, no extra text). The system will run the tool and reply with its result ` +
	`as a system message, after which you continue. Only call a tool when you actually need live data or to ` +
	`make a change; for general questions answer directly. When you have enough information, answer the user's ` +
	`original question in natural language.

Available tools:
`

// buildSystemPrompt assembles the chat system prompt, appending the descriptions
// of the MCP tools discovered on the server (if any).
func buildSystemPrompt(tools []utils.MCPTool) string {
	if len(tools) == 0 {
		return baseSystemPrompt
	}
	var b strings.Builder
	b.WriteString(baseSystemPrompt)
	b.WriteString(toolProtocolPrompt)
	for _, tool := range tools {
		b.WriteString(fmt.Sprintf("- %s: %s\n  input schema: %s\n", tool.Name, tool.Description, string(tool.InputSchema)))
	}
	return b.String()
}

var ChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session with Symphony",
	Long: `Start an interactive chat session between "User" and "Symphony".

The conversation is routed through the Symphony API's model router endpoint,
which forwards it to a configured OpenAI-compatible model. The assistant can also
call tools exposed by the Symphony MCP server to inspect and manage live objects;
those tool calls run under your identity. Type 'exit' or 'quit' (or press Ctrl+D)
to end the session.`,
	Run: func(cmd *cobra.Command, args []string) {
		c := config.GetMaestroConfig(configFile)
		ctx := c.DefaultContext
		if configContext != "" {
			ctx = configContext
		}
		if ctx == "" {
			ctx = "default"
		}

		mctx, ok := c.Contexts[ctx]
		if !ok {
			fmt.Printf("\n%s  configuration context '%s' is not found%s\n\n", utils.ColorRed(), ctx, utils.ColorReset())
			return
		}

		fmt.Printf("\n%sSymphony chat%s - talking to '%s'. Type %sexit%s or %squit%s to leave.\n\n",
			utils.ColorBlue(), utils.ColorReset(), ctx,
			utils.ColorYellow(), utils.ColorReset(),
			utils.ColorYellow(), utils.ColorReset())

		// Discover the MCP tools so the assistant can inspect and manage objects.
		tools, err := utils.MCPListTools(mctx.Url, mctx.User, mctx.Secret)
		if err != nil {
			fmt.Printf("%s  MCP tools are unavailable (%s); continuing without them.%s\n\n",
				utils.ColorYellow(), err.Error(), utils.ColorReset())
		} else if len(tools) > 0 {
			names := make([]string, 0, len(tools))
			for _, t := range tools {
				names = append(names, t.Name)
			}
			fmt.Printf("%s  MCP tools available: %s%s\n\n", utils.ColorYellow(), strings.Join(names, ", "), utils.ColorReset())
		}

		messages := []utils.ChatMessage{
			{Role: "system", Content: buildSystemPrompt(tools)},
		}

		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Printf("%sUser>%s ", utils.ColorGreen(), utils.ColorReset())
			line, err := reader.ReadString('\n')
			if err != nil {
				// EOF (Ctrl+D) or read error - end the session.
				fmt.Println()
				break
			}
			input := strings.TrimSpace(line)
			if input == "" {
				continue
			}
			if strings.EqualFold(input, "exit") || strings.EqualFold(input, "quit") {
				break
			}

			messages = append(messages, utils.ChatMessage{Role: "user", Content: input})
			updated, reply, err := chatWithTools(mctx, messages)
			if err != nil {
				fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
				// Drop the failed user turn so the conversation history stays consistent.
				messages = messages[:len(messages)-1]
				continue
			}
			messages = updated

			fmt.Printf("\n%sSymphony>%s %s\n\n", utils.ColorBlue(), utils.ColorReset(), reply)
		}

		fmt.Printf("\n%sGoodbye!%s\n\n", utils.ColorGreen(), utils.ColorReset())
	},
}

// chatWithTools sends the conversation to the model and transparently handles any
// "#TOOL <json>" directives the assistant emits: it invokes the requested tool on
// the Symphony MCP server (under the caller's identity), feeds the result back
// into the conversation, and repeats until the assistant returns a normal answer.
// It returns the updated message history and the final assistant reply.
func chatWithTools(mctx config.MaestroContext, messages []utils.ChatMessage) ([]utils.ChatMessage, string, error) {
	for step := 0; step < maxToolSteps; step++ {
		reply, err := utils.ChatCompletion(mctx.Url, mctx.User, mctx.Secret, chatEndpoint, chatModel, messages)
		if err != nil {
			return messages, "", err
		}
		messages = append(messages, utils.ChatMessage{Role: "assistant", Content: reply})

		name, arguments, ok := parseToolDirective(reply)
		if !ok {
			return messages, reply, nil
		}

		var content string
		result, cErr := utils.MCPCallTool(mctx.Url, mctx.User, mctx.Secret, name, arguments)
		if cErr != nil {
			content = fmt.Sprintf("Tool '%s' failed: %s", name, cErr.Error())
		} else {
			content = fmt.Sprintf("Tool '%s' result:\n%s", name, result)
		}
		messages = append(messages, utils.ChatMessage{Role: "system", Content: content})
	}
	return messages, "", errors.New("the assistant exceeded the maximum number of tool steps")
}

// parseToolDirective returns the tool name and arguments requested by a
// "#TOOL <json>" directive if the assistant reply is exactly such a directive.
func parseToolDirective(reply string) (string, map[string]interface{}, bool) {
	firstLine := strings.TrimSpace(reply)
	if idx := strings.IndexByte(firstLine, '\n'); idx >= 0 {
		firstLine = strings.TrimSpace(firstLine[:idx])
	}
	const prefix = "#TOOL "
	if len(firstLine) <= len(prefix) || !strings.EqualFold(firstLine[:len(prefix)], prefix) {
		return "", nil, false
	}
	jsonPart := strings.TrimSpace(firstLine[len(prefix):])
	var directive struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal([]byte(jsonPart), &directive); err != nil || directive.Name == "" {
		return "", nil, false
	}
	return directive.Name, directive.Arguments, true
}

func init() {
	ChatCmd.Flags().StringVarP(&chatModel, "model", "m", "gpt-4o", "The model to use for the chat session")
	ChatCmd.Flags().StringVarP(&chatEndpoint, "endpoint", "e", "", "The model router endpoint to use (defaults to the server's default endpoint)")
	ChatCmd.Flags().StringVarP(&configFile, "config", "c", "", "Maestro CLI config file")
	ChatCmd.Flags().StringVarP(&configContext, "context", "", "", "Maestro CLI configuration context")
	RootCmd.AddCommand(ChatCmd)
}
