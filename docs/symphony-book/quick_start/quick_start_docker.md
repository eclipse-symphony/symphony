# Use Symphony in a Docker container

_(last edit: 9/18/2023)_

You can run the Symphony API as a single Docker container with a configuration file that you mount to the container.

```bash
# assuming you are under the repo root folder
docker run --rm -it -e LOG_LEVEL=Info -v ./api:/configs -e CONFIG=/configs/symphony-api-no-k8s.json possprod.azurecr.io/symphony-api:latest
```

> **Pre-release NOTE**: ```possprod.azurecr.io``` is a private repo. To access the repo, your Azure account needs to be granted access. Then, you need to login to Docker using Azure token:
>
>```bash
>az login
>TOKEN=$(az acr login --name possprod --expose-token --output tsv --query accessToken)
>docker login possprod.azurecr.io --username 00000000-0000-0000-0000-000000000000 --password $TOKEN
>```

Now that you have SYmphony running with Docker, you can use REST endpoints to interact with the Symphony API.
