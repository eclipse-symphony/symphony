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
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/azure/symphony/api/constants"
	mu "github.com/azure/symphony/api/pkg/apis/v1alpha1/managers"
	spf "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers"
	svf "github.com/azure/symphony/api/pkg/apis/v1alpha1/vendors"
	host "github.com/azure/symphony/coa/pkg/apis/v1alpha2/host"
	mf "github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	pf "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providerfactory"
	vf "github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
	logger "github.com/azure/symphony/coa/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	configFile string
	logLevel   string
)

var RootCmd = &cobra.Command{
	Use:   "symphony-api",
	Short: "Symphony API",
	Long: `
	
	S Y M P H O N Y
	
	`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonFile, err := os.Open(configFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer jsonFile.Close()
		bytes, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		config := host.HostConfig{}
		err = json.Unmarshal(bytes, &config)
		if err != nil {
			fmt.Println(err)
			return
		}
		starHost := host.APIHost{}
		err = starHost.Launch(config, []vf.IVendorFactory{
			svf.SymphonyVendorFactory{},
		}, []mf.IManagerFactroy{
			&mu.SymphonyManagerFactory{},
		}, []pf.IProviderFactory{
			spf.SymphonyProviderFactory{},
		}, true)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func Execute(versiong string) {
	fmt.Println(constants.EulaMessage)
	fmt.Println()
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	defaultConfig := "symphony-api.json"
	user, err := user.Current()
	if err == nil {
		homeDirectory := user.HomeDir
		defaultConfig = filepath.Join(homeDirectory, defaultConfig)
	}
	RootCmd.Flags().StringVarP(&configFile, "config", "c", defaultConfig, "Symphony API configuration file")
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "Fatal", "set log level")
}

func initConfig() {
	loggerOptions := logger.DefaultOptions()
	err := loggerOptions.SetOutputLevel(logLevel)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err = logger.ApplyOptionsToLoggers(&loggerOptions); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
