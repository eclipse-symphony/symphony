# Symphony agent

_(last edit: 9/18/2023)_

Symphony doesn’t mandate an agent to be installed on the targets it manages. Symphony prefers remote management interfaces when possible. On the other hand, if a target doesn’t support a remote management interface, a Symphony agent can be installed on the target so that Symphony control plane can manage it.

Symphony has two types of agents – a Symphony agent and a Polling agent. The Symphony agent uses the same binary as the Symphony API, with a different configuration file. A polling agent can be written in any programming language and talks to the Symphony control plane over HTTP/HTTPS. Symphony has a default polling agent implementation, named Piccolo, which is designed for tiny edge devices. The agent itself is about 4MB in size and requires about 430K memory to run.

## Related topics

* [Symphony agent](./symphony-agent.md)
* [Symphony polling agent (Piccolo)](./polling-agent.md)
