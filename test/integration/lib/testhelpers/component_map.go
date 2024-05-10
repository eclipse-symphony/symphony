package testhelpers

type Map = map[string]interface{}
type Array = []interface{}

// Todo: Switch over to symphony core types from the /k8s/api folder
var (
	ComponetsMap = map[string]ComponentSpec{
		// A simple chart that deploy a single pod, a configmap and a serviceaccount
		"simple-chart-1": {
			Name: "simple-chart-1",
			Properties: Map{
				"chart": Map{
					"repo":    "symphonycr.azurecr.io/simple-chart",
					"version": "0.3.0",
				},
			},
			Type: "helm.v3",
		},

		"simple-chart-1-nonexistent": {
			Name: "simple-chart-1",
			Properties: Map{
				"chart": Map{
					"repo":    "symphonycr.azurecr.io/simple-chart",
					"version": "0.3.0-nonexistent",
				},
			},
			Type: "helm.v3",
		},

		"simple-chart-1-with-values": {
			Name: "simple-chart-1",
			Properties: Map{
				"chart": Map{
					"repo":    "symphonycr.azurecr.io/simple-chart",
					"version": "0.3.0",
				},
				"values": Map{
					"configData": Map{
						"key": "value",
					},
				},
			},
			Type: "helm.v3",
		},

		// A simple chart that deploy a single pod, a configmap and a serviceaccount
		"simple-chart-2": {
			Name: "simple-chart-2",
			Properties: Map{
				"chart": Map{
					"repo":    "symphonycr.azurecr.io/simple-chart",
					"version": "0.3.0",
				},
			},
			Type: "helm.v3",
		},

		// A non-exisitent chart
		"simple-chart-2-nonexistent": {
			Name: "simple-chart-2",
			Properties: Map{
				"chart": Map{
					"repo":    "symphonycr.azurecr.io/simple-chart",
					"version": "0.3.0-non-existent",
				},
			},
			Type: "helm.v3",
		},

		"mongodb-configmap": {
			Name: "mongodb",
			Type: "yaml.k8s",
			Properties: Map{
				"resource": Map{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": Map{
						"name": "mongodb",
					},
					"data": Map{
						"database":     "mongodb",
						"database_uri": "mongodb://localhost:27017",
					},
				},
			},
		},

		"mongodb-configmap-modified": {
			Name: "mongodb",
			Type: "yaml.k8s",
			Properties: Map{
				"resource": Map{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": Map{
						"name": "mongodb",
					},
					"data": Map{
						"database":     "mongodb",
						"database_uri": "mongodb://localhost:27020", // changed port
					},
				},
			},
		},

		"mongodb-constraint": {
			Name: "mongodb-constraint",
			Type: "yaml.k8s",
			Properties: Map{
				"resource": Map{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": Map{
						"name": "mongodb-constraint",
					},
					"data": Map{
						"database":     "mongodb",
						"database_uri": "mongodb://localhost:27017",
					},
				},
			},
			Constraints: "${{$equal($property('OS'),'windows')}}",
		},

		"nginx": {
			Name: "nginx",
			Properties: Map{
				"resource": Map{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": Map{
						"name": "nginx",
					},
					"spec": Map{
						"replicas": int64(1),
						"selector": Map{
							"matchLabels": Map{
								"app": "nginx",
							},
						},
						"template": Map{
							"metadata": Map{
								"labels": Map{
									"app": "nginx",
								},
							},
							"spec": Map{
								"containers": []Map{
									{
										"image": "nginx:1.21",
										"name":  "nginx",
										"ports": Array{
											Map{"containerPort": int64(80)},
										},
									},
								},
							},
						},
					},
				},
			},
			Type: "yaml.k8s",
		},

		"basic-clusterrole": {
			Name: "basic-clusterrole",
			Properties: Map{
				"resource": Map{
					"apiVersion": "rbac.authorization.k8s.io/v1",
					"kind":       "ClusterRole",
					"metadata": Map{
						"name": "basic-clusterrole",
					},
					"rules": Array{
						Map{
							"apiGroups": Array{
								"apps",
							},
							"resources": Array{
								"deployments",
							},
							"verbs": Array{
								"get",
								"list",
								"watch",
								"create",
								"update",
							},
						},
					},
				},
			},
			Type: "yaml.k8s",
		},

		"basic-configmap-1": {
			Name: "basic-configmap-1",
			Properties: Map{
				"resource": Map{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": Map{
						"name": "basic-configmap-1",
					},
					"data": Map{
						"key": "value",
					},
				},
			},
			Type: "yaml.k8s",
		},

		"basic-configmap-1-modified": {
			Name: "basic-configmap-1",
			Properties: Map{
				"resource": Map{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": Map{
						"name": "basic-configmap-1",
					},
					"data": Map{
						"key": "value-modified",
					},
				},
			},
			Type: "yaml.k8s",
		},
		"basic-configmap-1-params": {
			Name: "basic-configmap-1",
			Properties: Map{
				"resource": Map{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": Map{
						"name": "basic-configmap-1",
					},
					"data": Map{
						"database":     "@{{ $param('database')}}",
						"database_uri": "@{{ $param('database_uri')}}",
					},
				},
			},
			Type: "yaml.k8s",
		},
		"basic-configmap-1-params-modified": {
			Name: "basic-configmap-1",
			Properties: Map{
				"resource": Map{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": Map{
						"name": "basic-configmap-1",
					},
					"data": Map{
						"uri": "${{ $param('test')}}",
					},
				},
			},
			Type: "yaml.k8s",
		},
		"foobar-crd": {
			Name: "foobar-crd",
			Properties: Map{
				"resource": Map{
					"apiVersion": "apiextensions.k8s.io/v1",
					"kind":       "CustomResourceDefinition",
					"metadata": Map{
						"name": "foobars.contoso.io",
					},
					"spec": Map{
						"group":   "contoso.io",
						"version": "v1",
						"scope":   "Namespaced",
						"names": Map{
							"plural":   "foobars",
							"singular": "foobar",
							"kind":     "FooBar",
							"shortNames": Array{
								"fb",
							},
						},
						"versions": Array{
							Map{
								"name":    "v1",
								"served":  true,
								"storage": true,
								"schema": Map{
									"openAPIV3Schema": Map{
										"type": "object",
										"properties": Map{
											"spec": Map{
												"type": "object",
												"properties": Map{
													"foo": Map{
														"type": "string",
													},
													"bar": Map{
														"type": "string",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"statusProbe": Map{
					"succeededValues": Array{"True"},
					"statusPath":      `$.status.conditions[?(@.type == "Established")].status`,
					"initialWait":     "5s",
				},
			},
			Type: "yaml.k8s",
		},
		"simple-foobar": {
			Name: "simple-foobar",
			Properties: Map{
				"resource": Map{
					"apiVersion": "contoso.io/v1",
					"kind":       "FooBar",
					"metadata": Map{
						"name": "simple-foobar",
					},
					"spec": Map{
						"foo": "foo",
						"bar": "bar",
					},
				},
			},
			Type: "yaml.k8s",
			Dependencies: []string{
				"foobar-crd",
			},
		},

		// A simple chart with a simple templated expression.
		"expressions-1": {
			Name: "expressions-1",
			Properties: Map{
				"chart": Map{
					"repo":    "symphonycr.azurecr.io/simple-chart",
					"version": "0.3.0",
				},
				"foo":                `@{{ $property("color") + ' ' + $property("OS") }}`,
				"testGtNumbers":      `@{{ $gt("2", 1.0)}}`,
				"testGeNumbers":      `@{{ $ge(2, "1.0")}}`,
				"testLtNumbers":      `@{{ $lt("2", 1.0)}}`,
				"testLeNumbers":      `@{{ $le(2, "1.0")}}`,
				"testBetweenNumbers": `@{{ $between(2, "1", 3)}}`,
			},
			Type: "helm.v3",
		},

		// A simple chart with an invalid templated expression, $property("will-fail") does not exist on the target.
		"expressions-1-failed": {
			Name: "expressions-1",
			Properties: Map{
				"chart": Map{
					"repo":    "symphonycr.azurecr.io/simple-chart",
					"version": "0.3.0",
				},
				"foo":                `${{ $property("will-fail") + ' ' + $property("OS") }}`,
				"testGtNumbers":      `${{ $gt("2", 1.0)}}`,
				"testGeNumbers":      `${{ $ge(2, "1.0")}}`,
				"testLtNumbers":      `${{ $lt("2", 1.0)}}`,
				"testLeNumbers":      `${{ $le(2, "1.0")}}`,
				"testBetweenNumbers": `${{ $between(2, "1", 3)}}`,
			},
			Type: "helm.v3",
		},

		// A simple chart with a simple templated expression.
		"expressions-1-soln": {
			Name: "expressions-1-soln",
			Properties: Map{
				"resource": Map{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": Map{
						"name":             "expressions-1-soln",
						"foo":              `@{{ $property("color") + ' ' + $property("OS") }}`,
						"normalString":     `This is interpreted as a normal string @{{ $property("wont-fail") }}`,
						"testEqualNumbers": `@{{ $equal(123, 123) }}`,
						"testNotTrue":      `@{{ $not("true")}}`,
						"testNotNotTrue":   `@{{ $not($not(true))}}`,
						"testPropertyAnd":  `@{{ $and($equal($property("OS"), "windows") , $equal("yes", "no"))}}`,
						"testPropertyOr":   `@{{ $or($equal($property("OS"), "windows") , $equal("yes", "no"))}}`,
					},
				},
			},
			Type: "yaml.k8s",
		},

		// A simple chart with an invalid templated expression, $property("will-fail") does not exist on the target.
		"expressions-1-soln-failed": {
			Name: "expressions-1-soln",
			Properties: Map{
				"chart": Map{
					"repo":    "symphonycr.azurecr.io/simple-chart",
					"version": "0.3.0",
				},
				"name":             "expressions-1-soln",
				"foo":              `@{{ $property("will-fail") + ' ' + $property("OS") }}`,
				"normalString":     `This is interpreted as a normal string @{{ $property("wont-fail") }}`,
				"testEqualNumbers": `@{{ $equal(123, 123) }}`,
				"testNotTrue":      `@{{ $not("true")}}`,
				"testNotNotTrue":   `@{{ $not($not(true))}}`,
				"testPropertyAnd":  `@{{ $and($equal($property("OS"), "windows") , $equal("yes", "no"))}}`,
				"testPropertyOr":   `@{{ $or($equal($property("OS"), "windows") , $equal("yes", "no"))}}`,
			},
			Type: "helm.v3",
		},

		"simple-http": {
			Name: "simple-http",
			Properties: Map{
				"http.url":    "https://learn.microsoft.com/en-us/content-nav/azure.json?",
				"http.method": "GET",
			},
			Type: "http",
		},
		"simple-http-invalid-url": {
			Name: "simple-http",
			Properties: Map{
				"http.url":    "https://learn.microsoft.com/en-us/test/invalid/url",
				"http.method": "GET",
			},
			Type: "http",
		},
		"e4k": {
			Name: "e4k",
			Properties: map[string]interface{}{
				"chart": map[string]interface{}{
					"repo":    "e4kpreview.azurecr.io/helm/az-e4k",
					"version": "0.3.0",
				},
			},
			Type: "helm.v3",
		},
		"e4k-broker": {
			Name: "e4k-high-availability-broker",
			Properties: map[string]interface{}{
				"chart": map[string]interface{}{
					"repo":    "symphonycr.azurecr.io/az-e4k-broker",
					"version": "0.1.0",
				},
			},
			Type: "helm.v3",
		},
		"bluefin-extension": {
			Name: "bluefin",
			Properties: map[string]interface{}{
				"chart": map[string]interface{}{
					"repo":    "azbluefin.azurecr.io/helm/bluefin-arc-extension",
					"version": "0.2.0-20230706.3-develop",
				},
			},
			Type: "helm.v3",
		},
		"bluefin-instance": {
			Name: "bluefin-instance",
			Properties: map[string]interface{}{
				"resource": map[string]interface{}{
					"apiVersion": "bluefin.az-bluefin.com/v1",
					"kind":       "Instance",
					"metadata": map[string]interface{}{
						"name":      "bf-instance",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"displayName":          "Test Instance",
						"otelCollectorAddress": "otel-collector.alice-springs.svc.cluster.local:4317",
					},
				},
			},
			Type: "yaml.k8s",
		},

		"bluefin-pipeline": {
			Name: "test-pipeline",
			Properties: map[string]interface{}{
				"resource": map[string]interface{}{
					"apiVersion": "bluefin.az-bluefin.com/v1",
					"kind":       "Pipeline",
					"metadata": map[string]interface{}{
						"name":      "bf-pipeline",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"displayName": "bf-pipeline",
						"enabled":     true,
						"input": map[string]interface{}{
							"description": "Read from topic Thermostat 3",
							"displayName": "E4K",
							"format":      map[string]interface{}{"type": "json"},
							"mqttConnectionInfo": map[string]interface{}{
								"broker":   "tcp://azedge-dmqtt-frontend:1883",
								"password": "password",
								"username": "client1",
							},
							"next": []interface{}{"node-22f2"},
							"topics": []interface{}{
								map[string]interface{}{
									"name": "alice-springs/data/opc-ua-connector/opc-ua-connector/thermostat-sample-3",
								},
							},
							"type": "input/mqtt@v1",
							"viewOptions": map[string]interface{}{
								"position": map[string]interface{}{
									"x": 0,
									"y": 80,
								},
							},
						},
						"partitionCount": 6,
						"stages": map[string]interface{}{
							"node-22f2": map[string]interface{}{
								"displayName": "No-op",
								"next":        []interface{}{"output"},
								"query":       ".",
								"type":        "processor/transform@v1",
								"viewOptions": map[string]interface{}{
									"position": map[string]interface{}{
										"x": 0,
										"y": 208,
									},
								},
							},
							"output": map[string]interface{}{
								"broker":      "tcp://azedge-dmqtt-frontend:1883",
								"description": "Publish to topic demo-output-topic",
								"displayName": "E4K",
								"format":      map[string]interface{}{"type": "json"},
								"password":    "password",
								"timeout":     "45ms",
								"topic":       "alice-springs/data/demo-output",
								"type":        "output/mqtt@v1",
								"username":    "client1",
								"viewOptions": map[string]interface{}{
									"position": map[string]interface{}{
										"x": 0,
										"y": 336,
									},
								},
							},
						},
					},
				},
			},
			Type: "yaml.k8s",
		},
	}
)
