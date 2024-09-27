# Scenario: using private container registry in helm chart
In symphony context, instead of creating the iamge pull secret in the helm chart and passing the key and password in plain-text, you can pass the password using $secret to the chart values.

You can create a solution file like this:
The $secret refers to the kubenetes secret. $secret("SECRETNAME", "FIELDNAME")
The kubenetes secret can either be manually created or synced using secret management tools, like Fortos.
```
apiVersion: solution.symphony/v1
kind: SolutionContainer
metadata:
  name: test-app  
spec:
---
apiVersion: solution.symphony/v1
kind: Solution
metadata: 
  name: test-app-v-v1
spec:
  rootResource: test-app
  components:
  - name: some_chart_name
    properties:
      chart:
        repo: some_chart_repo
        name: some_chart_name
      values:
        imagePullSecrets:
            name: repo1Pulling
            username: $secret("repo1pullingsecret", "username")
            password: $secret("repo1pullingsecret", "pwd")
            repo: <PlaceHolder>
    type: helm.v3
```

Inside the chart, there would be a secret yaml to create the image pulling secret:
```
apiVersion: v1
kind: Secret
metadata:
  name: {{ .Values.imagePullSecrets.name }}
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: {{ printf "{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\",\"auth\":\"%s\"}}}" .Values.imagePullSecrets.repo .Values.imagePullSecrets.username .Values.imagePullSecrets.password (printf "%s:%s" .Values.imagePullSecrets.username .Values.imagePullSecrets.password | b64enc) | b64enc | quote }}
```
The .dockerconfigjson has a format like this:
```
'{"auths":{"https://index.docker.io/v1/":{"username":"<USERNAME>","password":"<PASSWORD>","email":"<EMAIL>","auth":"<USERNAME>:<PASSWORD>(base64)"}}}'
```
Email is optional

With this imagePullingSecret, you can create a deployment spec pulling images from the private container registry
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}
  labels:
    app: {{ .Chart.Name }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app: {{ .Chart.Name }}
  template:
    metadata:
      labels:
        app: {{ .Chart.Name }}
    spec:
      imagePullSecrets:
        - name: {{ .Values.imagePullSecrets.name }}
      containers:
        - name: {{ .Chart.Name }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          ports:
            - name: http
              containerPort: 80
              protocol: TCP
```

You can find a sample chart here:  
[sample-chart Chart.yaml](../../samples/privateContainerRegistry/Chart.yaml)