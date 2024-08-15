/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/cli/config"
	"github.com/eclipse-symphony/symphony/cli/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var (
	setSwitches         []string
	sampleConfigFile    string
	sampleConfigContext string
)
var SamplesCmd = &cobra.Command{
	Use:   "samples",
	Short: "Symphony samples",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\n%sPlease use either 'samples run' or 'samples list'%s\n\n", utils.ColorRed(), utils.ColorReset())
	},
}
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "list Symphony samples",
	Run: func(cmd *cobra.Command, args []string) {
		sampleManifest, err := listSamples()
		if err != nil {
			fmt.Printf("\n%s%s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		if len(sampleManifest.Samples) == 0 {
			fmt.Printf("\n%sNo samples found%s\n\n", utils.ColorRed(), utils.ColorReset())
			return
		}
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Name", "Description", "Requires"})
		for _, sample := range sampleManifest.Samples {
			row := table.Row{sample.Name, sample.Description, strings.Join(sample.Requires, ", ")}
			t.AppendRow(row)
		}
		t.SetStyle(table.StyleColoredBright)
		t.Render()
	},
}

type SampleManifest struct {
	Package string       `json:"package"`
	Version string       `json:"version"`
	Samples []SampleSpec `json:"samples"`
}
type SampleSpec struct {
	Name            string         `json:"name"`
	Path            string         `json:"path"`
	Artifacts       []ArtifactSpec `json:"artifacts"`
	Description     string         `json:"description"`
	LongDescription string         `json:"description-long,omitempty"`
	Requires        []string       `json:"requires"`
	PostActions     []ActionSpec   `json:"postActions,omitempty"`
}
type ArtifactSpec struct {
	File       string          `json:"file"`
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	Parameters []ParameterSpec `json:"parameters,omitempty"`
}
type ParameterSpec struct {
	Name    string `json:"name"`
	Value   string `json:"value,omitempty"`
	Replace string `json:"replace"`
}
type ActionSpec struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

func listSamples() (SampleManifest, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	sampleManifest := filepath.Join(dirname, ".symphony", "samples.json")
	jsonFile, err := os.Open(sampleManifest)
	if err != nil {
		return SampleManifest{}, fmt.Errorf("sample manifest file '%s' is not found", sampleManifest)
	}
	defer jsonFile.Close()
	data, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return SampleManifest{}, fmt.Errorf("failed to read sample manifest: %s", err.Error())
	}
	var ret SampleManifest
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return SampleManifest{}, fmt.Errorf("failed to parse sample manifest: %s", err.Error())
	}
	return ret, nil
}

var DescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "describe a Symphony sample",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("\n%sPlease specify one sample name%s\n\n", utils.ColorRed(), utils.ColorReset())
			return
		}
		sampleManifest, err := listSamples()
		if err != nil {
			fmt.Printf("\n%s%s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		if len(sampleManifest.Samples) == 0 {
			fmt.Println("no samples found")
			fmt.Printf("\n%s  No samples found. %s\n\n", utils.ColorRed(), utils.ColorReset())
			return
		}
		for _, sample := range sampleManifest.Samples {
			if sample.Name == args[0] {
				if sample.LongDescription != "" {
					str := strings.ReplaceAll(sample.LongDescription, "<C>", utils.ColorCyan())
					str = strings.ReplaceAll(str, "</C>", utils.ColorReset())
					str = strings.ReplaceAll(str, "<G>", utils.ColorGreen())
					str = strings.ReplaceAll(str, "</G>", utils.ColorReset())
					fmt.Printf("\n%s\n\n", str)
					return
				}
				fmt.Printf("\n%s\n\n", sample.Description)
				return
			}
		}
		fmt.Printf("\n%sSample '%s' is not found, please use maestro samples list to check available samples%s\n", utils.ColorRed(), args[0], utils.ColorReset())
	},
}

var SampleRunCmd = &cobra.Command{
	Use:   "run",
	Short: "run a Symphony sample",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("\n%sPlease specify one sample name%s\n\n", utils.ColorRed(), utils.ColorReset())
			return
		}
		sampleManifest, err := listSamples()
		if err != nil {
			fmt.Printf("\n%s%s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		if len(sampleManifest.Samples) == 0 {
			fmt.Println("no samples found")
			fmt.Printf("\n%s  No samples found. %s\n\n", utils.ColorRed(), utils.ColorReset())
			return
		}
		for _, sample := range sampleManifest.Samples {
			if sample.Name == args[0] {
				paramMap, err := setSwitchToMap()
				if err != nil {
					fmt.Println(err)
					return
				}
				for i, artifact := range sample.Artifacts {
					err := runArtifact(sample.Path, artifact, paramMap)
					if err != nil {
						fmt.Printf("\n%s  Failed to run sample: %s %s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
						return
					}
					if i < len(sample.Artifacts)-1 {
						time.Sleep(3 * time.Second)
					}
				}
				xArg := ""
				for _, action := range sample.PostActions {
					if xArg != "" {
						for i, _ := range action.Args {

							action.Args[i] = strings.ReplaceAll(action.Args[i], "$(1)", strings.Trim(xArg, `'"`))
						}
					}
					errOutput := ""
					xArg, errOutput, err = utils.RunCommandWithRetry("", "", verbose, debug, action.Command, action.Args...)
					if err != nil {
						fmt.Printf("\n%s  Failed to execute post deployment script: %s %s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
						fmt.Printf("\n%s  Detailed Messages: %s %s\n\n", utils.ColorRed(), errOutput, utils.ColorReset())
						return
					}
				}
				if xArg != "" {
					fmt.Printf("  %sNOTE%s: %s\n", utils.ColorGreen(), utils.ColorReset(), xArg)
				}
				return
			}
		}
		fmt.Printf("\n%sSample '%s' is not found, please use maestro samples list to check available samples%s\n", utils.ColorRed(), args[0], utils.ColorReset())
	},
}

func setSwitchToMap() (map[string]string, error) {
	ret := make(map[string]string)
	for _, s := range setSwitches {
		i := strings.Index(s, "=")
		if i <= 0 {
			return nil, fmt.Errorf("invalid parameter format '%s', expected key=value", s)
		}
		ret[s[:i]] = s[i+1:]
	}
	return ret, nil
}
func removeArtifact(artifact ArtifactSpec) error {
	c := config.GetMaestroConfig(sampleConfigFile)
	ctx := c.DefaultContext
	if sampleConfigContext != "" {
		ctx = sampleConfigContext
	}

	if ctx == "" {
		ctx = "default"
	}

	fmt.Printf("%sRemoving %s %s%s ...", utils.ColorCyan(), artifact.Type, utils.ColorReset(), artifact.Name)
	err := utils.Remove(
		c.Contexts[ctx].Url,
		c.Contexts[ctx].User,
		c.Contexts[ctx].Secret,
		artifact.Type,
		artifact.Name)
	if err != nil {
		fmt.Printf("%sfailed\n%s", utils.ColorRed(), utils.ColorReset())
		return err
	}
	fmt.Printf("%sdone\n%s", utils.ColorGreen(), utils.ColorReset())
	time.Sleep(8 * time.Second)
	return nil
}
func runArtifact(samplePath string, artifact ArtifactSpec, paramMap map[string]string) error {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
		return err
	}
	artifactFile := filepath.Join(dirname, ".symphony", samplePath, artifact.File)
	if _, err := os.Stat(artifactFile); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("sample artifact '%s' is not found", artifactFile)
	}
	src, err := os.ReadFile(artifactFile)
	strStr := string(src)
	if len(artifact.Parameters) > 0 {
		if err != nil {
			return err
		}
		for _, p := range artifact.Parameters {
			if v, ok := paramMap[p.Name]; ok {
				p.Value = v
			}
			if p.Value == "" {
				return fmt.Errorf("parameter value '%s' is not set", p.Name)
			}
			strStr = strings.ReplaceAll(strStr, p.Replace, p.Value)
		}
	}
	c := config.GetMaestroConfig(sampleConfigFile)
	ctx := c.DefaultContext
	if sampleConfigContext != "" {
		ctx = sampleConfigContext
	}

	if ctx == "" {
		ctx = "default"
	}

	fmt.Printf("%sCreating %s %s%s ... ", utils.ColorCyan(), artifact.Type, utils.ColorReset(), artifact.Name)

	err = utils.Upsert(
		c.Contexts[ctx].Url,
		c.Contexts[ctx].User,
		c.Contexts[ctx].Secret,
		artifact.Type,
		artifact.Name,
		[]byte(strStr))
	if err != nil {
		fmt.Printf("%sfailed\n%s", utils.ColorRed(), utils.ColorReset())
		return err
	}
	fmt.Printf("%sdone\n%s", utils.ColorGreen(), utils.ColorReset())
	return nil
}

var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a Symphony sample",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("\n%sPlease specify one sample name%s\n\n", utils.ColorRed(), utils.ColorReset())
			return
		}
		sampleManifest, err := listSamples()
		if err != nil {
			fmt.Printf("\n%s%s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		if len(sampleManifest.Samples) == 0 {
			fmt.Println("no samples found")
			fmt.Printf("\n%s  No samples found. %s\n\n", utils.ColorRed(), utils.ColorReset())
			return
		}
		for _, sample := range sampleManifest.Samples {
			if sample.Name == args[0] {
				for i := len(sample.Artifacts) - 1; i >= 0; i-- {
					err := removeArtifact(sample.Artifacts[i])
					if err != nil {
						fmt.Printf("\n%s  Failed to remove sample: %s %s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
						return
					}
					if i > 0 {
						time.Sleep(3 * time.Second)
					}
				}
				return
			}
		}
		fmt.Printf("Sample '%s' is not found, please use maestro samples list to check available samples\n", args[0])
	},
}

func init() {
	SampleRunCmd.Flags().StringArrayVarP(&setSwitches, "set", "s", nil, "set sample parameter as key=value")
	SampleRunCmd.Flags().StringVarP(&sampleConfigFile, "config", "c", "", "Maestro CLI config file")
	SampleRunCmd.Flags().StringVarP(&sampleConfigContext, "context", "", "", "Maestro CLI configuration context")
	RemoveCmd.Flags().StringVarP(&sampleConfigFile, "config", "c", "", "Maestro CLI config file")
	RemoveCmd.Flags().StringVarP(&sampleConfigContext, "context", "", "", "Maestro CLI configuration context")
	DescribeCmd.Flags().StringVarP(&sampleConfigFile, "config", "c", "", "Maestro CLI config file")
	DescribeCmd.Flags().StringVarP(&sampleConfigContext, "context", "", "", "Maestro CLI configuration context")
	SamplesCmd.AddCommand(SampleRunCmd)
	SamplesCmd.AddCommand(RemoveCmd)
	SamplesCmd.AddCommand(DescribeCmd)
	SamplesCmd.AddCommand(ListCmd)
	RootCmd.AddCommand(SamplesCmd)
}
