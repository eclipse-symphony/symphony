# Day 1: Hello, Symphony!

Meanwhile, in a parallel universe, Zonic Cars is having a team meeting in their Lisbon headquarter...

![Headquarter](../images/headquater.jpg)

|||
|----|----|
|![zonic](../images/zonic-small.png)|Our business is booming, which is good news! However, as our global footprint increases, we are facing a lot of consistency issues. Each site is managed somewhat differently, and we have all kinds of systems accumulated over the years. I often don’t even know what we have on some of the sites! I think this is getting out of hand, don’t you think? |
|![danny](../images/danny-small.png)|Yeah, there are definitely some repetitive work that has to be carried out manually on each site, which is inefficient and error-prone. We have configurations scattered everywhere, and nobody knows what configurations are applied and why. |
|![bob](../images/bob-small.png)|I agree. But we also need to give the sites more flexibility in some cases. Like when I’m not ready to take a new software version I should be able to reject or postpone it. And when I changed something on my site, global IT shouldn’t just come in and override it from the cloud. It has caused a few outages in our product lines.|
|![anna](../images/anna-small.png)| Some critical fixes must be pushed out as quickly as possible. I don’t think a site has a saying in applying critical patches. |
|![bob](../images/bob-small.png)| Well, certainly not when I’m about to weld a car frame! No one can touch my EPP-300 system, unless I approve. |
|![anna](../images/anna-small.png)| EPP-300? I thought we were switching to  AK-500 already. |
|![danny](../images/danny-small.png)|Yeah, but that will take at least a few years. So, we need to support both. |
|![anna](../images/anna-small.png)| But AK-500 works very differently from EPP-300. It has completely different APIs.|
|![zonic](../images/zonic-small.png)|It doesn’t matter if it’s EPP-300 or AK-500, we need to figure out, and implement our standardized business practices regardless of what machines we use. And I know we’ll be using a mixture of different things for sure, not just for ourselves but also for our partners. I feel like we have all the wonderful instruments in our hands. And with effort, they all play nice tunes. But what I want is a **Symphony**, in which all the instruments work in harmony to produce a greater value. |
|![george](../images/george-small.png)| Symphony, you say... I may have the exact thing you need! There’s a project called Symphony. It’s an orchestrator of tools and services to form consistent workflows in a distributed, heterogeneous edge infrastructure. And it only takes seconds to set it up. Let me show you…|

## Exercise 1: Bootstrap Symphony
**George:** Symphony runs in either a standalone mode as a single binary, or a Kubernetes model where Symphony runs as part of the Kubernetes API server. To deploy in standalone mode, you can just copy the single binary file to your machine - Linux, Windows or Mac, and launch it with a configuration file. No external depedencies or packages to install. To deploy on Kubernetes, you simply apply a Helm chart.

**George:** There's also a command-line tool, called **maestro**, which can help you to get started with Symphony quite easily. To install maestro, use this one-liner:

### On Linux/Mac

```bash
wget -q https://raw.githubusercontent.com/Haishi2016/Vault818/master/cli/install/install.sh -O - | /bin/bash
```
### On Windows
```bash
powershell -Command "iwr -useb https://raw.githubusercontent.com/Haishi2016/Vault818/master/cli/install/install.ps1 | iex"
```
> NOTE: The Haishi2016 repo is a temporary parking repo, which will be replaced before release.

**George:** Once you have maestro installed, you can launch Symphony in standalone mode by:
```
maestro up --no-k8s
```
**George:** It displays logs in the terminal window. Because everything is kept in memory, the system is wiped clean when you stop the process. This is perfect for dev/test purposes with zero configuration and zero depedencies. Now you can use any HTTP client to access Symphony REST API. For example, in a different terminal window, you can set a GET request to a test endpoint:

```bash
curl http://localhost:8082/v1alpha2/greetings
```

**George:** This is a test endpoint without authentication, so that you can easily check if the server is up and running. You'll get some text like:
```bash
Hello from Symphony K8s control plane (S8C)
```
**George:** And there you go! You just deployed Symphony in a standalone mode! It's as easy as that!

|||
|----|----|
|![anna](../images/anna-small.png)| That’s easy! But what can you do with it? |
|![george](../images/george-small.png)| This is a single machine deployment. You can imagine this Symphony is managing a site of one machine. Now, I can register my machine with Symphony as a **Target**. Then, I can start to deploy **Solution**s to it.|
|![anna](../images/anna-small.png)| **Solution** is like an application? |
|![george](../images/george-small.png)| Exactly. A **Solution** consists of one more multiple **Component**s, like Docker containers in a typical microservice archtiecture.|
|![anna](../images/anna-small.png)| Does it have to be containers? |
|![george](../images/george-small.png)| Not at all. Symphony support various component types and it's extensible to support more types, like binaries, app packages and OS images. You can also define depedencies among components so that they are installed in the correct order. |
|![anna](../images/anna-small.png)| So it handles containerized applications as well as "classic" applications. |
|![george](../images/george-small.png)| That's the idea - consistency regardless of application type or architecture. |
|![anna](../images/anna-small.png)| Got it! So, **Solution** is your application and **Target** is your machine. How do you create a deployment? |
|![george](../images/george-small.png)| You create a deployment by creating an **Instance** object. Basically, an **Instance** object defines which **Solution** should be put on which **Target**s. Let me show you.|
## Exercise 2: Deploying a Docker container
**George:** Now, I'm going to deploy a Redis server as a Docker container on my machine. This happens to be one of the sample scenarios shipped with **maestro**. To create the component trio: **Solution**, **Target** and **Instance**, I can simply do:
```bash
maestro samples run redis-docker
```
**George:** And if I do ```docker ps``` now, I can see a new Redis container got launched a few seconds ago. That's it!
**George:** And because Symphony does continuous state seeking, if the container is shut down for any reason, Symphony will bring it back. Now, I'm going to forcely shut down the container:
```bash
docker rm -f redis-server
```
**George:** But in a few seconds, if I do ```docker ps``` again, you can see the container is brought back, because my desired state in my **Instance** says I should have such container running on my machine.

**Anna**: This is pretty cool! But this assumes you have Docker on your machine, right?

**George:** Right, Symphony is an orchestrator. It doesn't aim to replace any of the technologies. It's focus is to create a consistent workflow on top of different technologies. For example, with Symphony, you use the same workflow regardless if you are deploying containers, installing application packages, or even applying OS images.

**Anna**: Awesome!

|||
|----|----|
|![danny](../images/danny-small.png)|But wait, we can't run in production as a single process, though...
|![george](../images/george-small.png)| Right. The standalone mode is for local dev/test only. For a production deployment, you probably want to deploy Symphony onto a Kubernetes cluster. Symphony runs natively on Kubernetes by extending the Kubernete's API server. You can install Symphony to popular Kubernetes distributions using Helm.|
## Exercise 3: Deploy Symphony to Kubernetes
**George:** Like I've mentioned, you can install Symphony using Helm. Or, **maestro** allows you to install Symphony to your current Kubernetes context by:
```
maestro up
```
**George:** And Symphony will be configured on your Kurbernetes cluster!

**George:** Once Symphony is installed, you can see it installs a few custom resource types, or [CRDs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/), including the resource types we've talked about - like **Solution**, **Target** and **Instance**. If you do:
```bash
kubectl get crds | grep symphony
```
**George:** You can see a list of Symphony resource types like:
```bash
activations.workflow.symphony 
campaigns.workflow.symphony
devices.fabric.symphony
instances.solution.symphony
models.ai.symphony
skillpackages.ai.symphony
skills.ai.symphony
solutions.solution.symphony
targets.fabric.symphony
```
**George:** There are certainly a lot more than the basic deployments we've done! But we can go through these later...

**George:** Meanwhile, let me just deploy another sample, which runs [Prometheus](https://prometheus.io/). Again, I'll just use the **maestro** tool:
```bash
maestro samples run hello-k8s
``` 

