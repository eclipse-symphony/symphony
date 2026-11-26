/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package verify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/require"
)

func TestScenarioResourceIsolation(t *testing.T) {
	executable, err := os.Executable()
	require.NoError(t, err)
	for _, namespace := range []string{"default", "nondefault"} {
		t.Run(namespace, func(t *testing.T) {
			objects := map[string]string{
				"/apis/fabric.symphony/v1/namespaces/" + namespace + "/targets/configmap-target":         `{"apiVersion":"fabric.symphony/v1","kind":"Target","metadata":{"name":"configmap-target"},"status":{"provisioningStatus":{"status":"Succeeded"}}}`,
				"/apis/solution.symphony/v1/namespaces/" + namespace + "/instances/instanceforconfigmap": `{"apiVersion":"solution.symphony/v1","kind":"Instance","metadata":{"name":"instanceforconfigmap"},"status":{"provisioningStatus":{"status":"Succeeded"}}}`,
				"/apis/workflow.symphony/v1/namespaces/" + namespace + "/activations/05activation":       fmt.Sprintf(`{"apiVersion":"workflow.symphony/v1","kind":"Activation","metadata":{"name":"05activation"},"status":{"status":%d,"stageHistory":[{"status":%d},{"status":%d,"inputs":{"age":10,"job":"worker","name":"sample"}}]}}`, v1alpha2.Done, v1alpha2.Done, v1alpha2.Done),
				"/api/v1/namespaces/default/configmaps/testconfigmap1":                                   `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"testconfigmap1"},"data":{"top":"noimage"}}`,
				"/api/v1/namespaces/default/configmaps/testconfigmap2":                                   `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"testconfigmap2"},"data":{"top":"noReplicas"}}`,
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if object, ok := objects[r.URL.Path]; ok {
					fmt.Fprint(w, object)
					return
				}
				for path, object := range objects {
					if filepath.Dir(path) == r.URL.Path {
						fmt.Fprintf(w, `{"apiVersion":"v1","kind":"List","items":[%s,{"apiVersion":"v1","kind":"Target","metadata":{"name":"unrelated"}}]}`, object)
						return
					}
				}
				http.Error(w, `{"message":"unexpected resource or namespace"}`, http.StatusNotFound)
			}))
			defer server.Close()

			home := t.TempDir()
			require.NoError(t, os.Mkdir(filepath.Join(home, ".kube"), 0700))
			config := map[string]interface{}{
				"apiVersion": "v1", "kind": "Config",
				"clusters":        []interface{}{map[string]interface{}{"name": "test", "cluster": map[string]interface{}{"server": server.URL}}},
				"contexts":        []interface{}{map[string]interface{}{"name": "test", "context": map[string]interface{}{"cluster": "test", "user": "test"}}},
				"users":           []interface{}{map[string]interface{}{"name": "test", "user": map[string]interface{}{}}},
				"current-context": "test",
			}
			data, err := json.Marshal(config)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(home, ".kube", "config"), data, 0600))
			t.Setenv("HOME", home)
			t.Setenv("NAMESPACE", namespace)
			cmd := exec.Command(executable, "-test.run=^TestBasic_(TargetStatus|InstanceStatus|ConfigmapStatus|ReadNamespaceActivation)$", "-test.timeout=10s")
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "%s", output)
		})
	}
}
