# Use Symphony in a Docker container

_(last edit: 9/18/2023)_

You can run the Symphony API as a single Docker container with a configuration file that you mount to the container.

```bash
# assuming you are under the repo root folder
docker run --rm -it -e LOG_LEVEL=Info -v ./api:/configs -e CONFIG=/configs/symphony-api-no-k8s.json ghcr.io/azure/symphony/symphony-api:latest
```

> **Pre-release NOTE**: ```ghcr.io``` is a private repo. To access the repo, you need to follow [this](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry#authenticating-to-the-container-registry) to generate the PAT token 
>
>```bash
>TOKEN='{YOUR_GITHUB_PAT_TOKEN}'
>docker login ghcr.io --username USERNAME --password $TOKEN
>```

Now that you have SYmphony running with Docker, you can use REST endpoints to interact with the Symphony API.
