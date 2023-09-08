/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"sort"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/cli/config"
	"github.com/azure/symphony/cli/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var (
	objectName    string
	configFile    string
	jsonPath      string
	docType       string
	configContext string
)
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Query Symphony objects",
	Run: func(cmd *cobra.Command, args []string) {
		c := config.GetMaestroConfig(configFile)
		ctx := c.DefaultContext
		if configContext != "" {
			ctx = configContext
		}

		if ctx == "" {
			ctx = "default"
		}

		for _, a := range args {
			list, err := utils.Get(
				c.Contexts[ctx].Url,
				c.Contexts[ctx].User,
				c.Contexts[ctx].Secret,
				a,
				jsonPath,
				docType,
				objectName)
			if err != nil {
				fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
				return
			}
			outputList(list, a, jsonPath)
		}
	},
}

func outputAsAttributes(t table.Writer, data []byte, objType string, path string, keys []string) error {
	var topAttrs map[string]interface{}
	err := json.Unmarshal(data, &topAttrs)
	if err != nil {
		return err
	}
	row := table.Row{}

	for _, key := range keys {
		v := topAttrs[key]
		if _, ok := v.(map[string]interface{}); ok {
			row = append(row, "map[...]")
		} else if _, ok := v.([]interface{}); ok {
			row = append(row, "array[...]")
		} else {
			row = append(row, v)
		}
	}
	t.AppendRow(row)
	return nil
}
func outputAsStr(t table.Writer, data []byte) error {
	var strVal string
	err := json.Unmarshal(data, &strVal)
	if err != nil {
		return err
	}
	t.AppendRow(table.Row{strVal})
	return nil
}
func outputAsArray(t table.Writer, item interface{}) error {
	if arr, ok := item.([]interface{}); ok {
		for _, a := range arr {
			row := table.Row{}
			if dict, ok := a.(map[string]interface{}); ok {
				for _, v := range dict {
					if _, ok := v.(map[string]interface{}); ok {
						row = append(row, "map[...]")
					} else {
						row = append(row, v)
					}
				}
			}
			t.AppendRow(row)
		}
	}
	return nil
}
func addTableHeader(t table.Writer, list interface{}, objType string, path string, itemType string) []string {
	if path == "" {
		switch objType {
		case "target", "targets":
			t.AppendHeader(table.Row{"Name", "Status"})
			return []string{"Name", "Status"}
		case "device", "devices":
			t.AppendHeader(table.Row{"Name", "Status"})
			return []string{"Name", "Status"}
		case "solution", "solutions":
			t.AppendHeader(table.Row{"Name"})
			return []string{"Name"}
		case "instance", "instances":
			t.AppendHeader(table.Row{"Name", "Status", "Targets", "Deployed"})
			return []string{"Name", "Status", "Targets", "Deployed"}
		}
		return nil
	}
	if itemType == "string" {
		header := path[strings.LastIndex(path, ".")+1:]
		t.AppendHeader(table.Row{header})
		return []string{header}
	}
	if itemType == "array" {
		arr := list.([]interface{})
		innerType := interfaceType(arr[0])
		return addTableHeader(t, arr[0], objType, path, innerType)
	}
	if itemType == "property-bag" {
		if dict, ok := list.(map[string]interface{}); ok {
			keys := make([]string, 0)
			for k, _ := range dict {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			row := table.Row{}
			for _, k := range keys {
				row = append(row, k)
			}
			t.AppendHeader(row)
			return keys
		}
	}
	return nil
}
func outputList(list []interface{}, objType string, path string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	if len(list) > 0 {
		itemType := interfaceType(list[0])
		keys := addTableHeader(t, list[0], objType, path, itemType)
		for _, item := range list {
			outputListItem(t, item, objType, path, keys)
		}
	}
	t.SetStyle(table.StyleColoredBright)
	t.Render()
}
func outputTarget(t table.Writer, data []byte) {
	var target Target
	err := json.Unmarshal(data, &target)
	if err == nil {
		row := table.Row{}
		row = append(row, target.Id)
		row = append(row, target.Status["status"])
		t.AppendRow(row)
	}
}
func outputDevice(t table.Writer, data []byte) {
	var device Device
	err := json.Unmarshal(data, &device)
	if err == nil {
		row := table.Row{}
		row = append(row, device.Id)
		row = append(row, device.Status["status"])
		t.AppendRow(row)
	}
}
func outputSolution(t table.Writer, data []byte) {
	var solution Solution
	err := json.Unmarshal(data, &solution)
	if err == nil {
		row := table.Row{}
		row = append(row, solution.Id)
		t.AppendRow(row)
	}
}
func outputInstance(t table.Writer, data []byte) {
	var instance Instance
	err := json.Unmarshal(data, &instance)
	if err == nil {
		row := table.Row{}
		row = append(row, instance.Id)
		row = append(row, instance.Status["status"])
		row = append(row, instance.Status["targets"])
		row = append(row, instance.Status["deployed"])
		t.AppendRow(row)
	}
}
func interfaceType(item interface{}) string {
	if _, ok := item.(map[string]interface{}); ok {
		return "property-bag"
	}
	if _, ok := item.(string); ok {
		return "string"
	}
	if _, ok := item.([]string); ok {
		return "string-array"
	}
	if _, ok := item.([]interface{}); ok {
		return "array"
	}
	return ""
}
func outputListItem(t table.Writer, item interface{}, objType string, path string, keys []string) {
	data, _ := json.Marshal(item)
	if path == "" {
		switch objType {
		case "device", "devices":
			outputDevice(t, data)
			return
		case "target", "targets":
			outputTarget(t, data)
			return
		case "solution", "solutions":
			outputSolution(t, data)
			return
		case "instance", "instances":
			outputInstance(t, data)
			return
		}
	}
	err := outputAsAttributes(t, data, objType, path, keys)
	if err == nil {
		return
	}
	err = outputAsStr(t, data)
	if err == nil {
		return
	}
	err = outputAsArray(t, item)
	if err == nil {
		return
	}
}

func init() {
	GetCmd.Flags().StringVarP(&objectName, "name", "n", "", "Symphony object name")
	GetCmd.Flags().StringVarP(&configFile, "config", "c", "", "Maestro CLI config file")
	GetCmd.Flags().StringVarP(&jsonPath, "json-path", "", "", "Jason Path query to be applied on results")
	GetCmd.Flags().StringVarP(&docType, "doc-type", "", "", "Result type (Json or Yaml)")
	GetCmd.Flags().StringVarP(&configContext, "context", "", "", "Maestro CLI configuration context")
	RootCmd.AddCommand(GetCmd)
}

type Target struct {
	Id     string            `json:"id"`
	Spec   model.TargetSpec  `json:"spec,omitempty"`
	Status map[string]string `json:"status,omitempty"`
}
type Device struct {
	Id     string            `json:"id"`
	Spec   model.DeviceSpec  `json:"spec,omitempty"`
	Status map[string]string `json:"status,omitempty"`
}
type Solution struct {
	Id     string             `json:"id"`
	Spec   model.SolutionSpec `json:"spec,omitempty"`
	Status map[string]string  `json:"status,omitempty"`
}
type Instance struct {
	Id     string             `json:"id"`
	Spec   model.InstanceSpec `json:"spec,omitempty"`
	Status map[string]string  `json:"status,omitempty"`
}
